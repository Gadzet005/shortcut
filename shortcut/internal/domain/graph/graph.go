package graph

import (
	"context"
	"errors"
	"sync"

	errorsutils "github.com/Gadzet005/shortcut/shortcut/pkg/utils/errors"
	"go.uber.org/zap"
)

type GraphID string

func (i GraphID) String() string {
	return string(i)
}

type Graph struct {
	ID    GraphID
	Nodes map[NodeID]Node
}

func (g Graph) Run(
	ctx context.Context,
	logger *zap.Logger,
	req RunNodeRequest,
) (RunNodeResponse, error) {
	results := req.Items

	levels := topSortByLevels(g)
	for _, level := range levels {
		err := visitNodes(ctx, logger, req, level, results)
		if err != nil {
			return RunNodeResponse{}, err
		}
	}

	lastNode := levels[len(levels)-1][0]
	lastNodeData, ok := results[ItemID{NodeID: lastNode.ID(), Name: DefaultItemName}]
	if !ok {
		return RunNodeResponse{}, errors.New("failed to get last node data")
	}
	return RunNodeResponse{Items: map[ItemName]Item{
		DefaultItemName: lastNodeData,
	}}, nil
}

// visitNodes посещает ноды паралельно и записывает ответы в results
func visitNodes(
	ctx context.Context,
	logger *zap.Logger,
	req RunNodeRequest,
	nodes []Node,
	results map[ItemID]Item,
) error {
	tmpResults := make([]map[ItemName]Item, len(nodes))
	nodeErrors := make([]error, len(nodes))

	var wg sync.WaitGroup
	wg.Add(len(nodes))
	for i, node := range nodes {
		go func() {
			defer wg.Done()

			result, err := visitNode(ctx, logger, req, node, results)
			if err != nil {
				nodeErrors[i] = err
			}
			tmpResults[i] = result
		}()
	}

	wg.Wait()

	for i, nodeError := range nodeErrors {
		if nodeError != nil {
			return errorsutils.WrapFail(nodeError, "visit node %s", nodes[i].ID())
		}
	}

	for i, result := range tmpResults {
		for name, item := range result {
			results[ItemID{NodeID: nodes[i].ID(), Name: name}] = item
		}
	}
	return nil
}

func visitNode(
	ctx context.Context,
	logger *zap.Logger,
	req RunNodeRequest,
	node Node,
	results map[ItemID]Item,
) (map[ItemName]Item, error) {
	items := make(map[ItemID]Item, len(node.Dependencies()))
	for _, dep := range node.Dependencies() {
		result, ok := results[dep]
		if !ok {
			return nil, errors.New("failed to get dependency")
		}
		items[dep] = result
	}

	resp, err := node.Run(ctx, logger, RunNodeRequest{
		Client: req.Client,
		Items:  items,
	})
	if err != nil {
		return nil, errorsutils.WrapFail(err, "run node %s", node.ID())
	}
	return resp.Items, nil
}

func topSortByLevels(g Graph) [][]Node {
	inDegree := make(map[NodeID]int)
	for _, node := range g.Nodes {
		if _, exists := inDegree[node.ID()]; !exists {
			inDegree[node.ID()] = 0
		}
		for _, dep := range node.Dependencies() {
			inDegree[dep.NodeID]++
		}
	}

	var levels [][]Node
	var currentLevel []Node
	for _, node := range g.Nodes {
		if inDegree[node.ID()] == 1 {
			currentLevel = append(currentLevel, node)
		}
	}

	for len(currentLevel) > 0 {
		levels = append(levels, currentLevel)
		var nextLevel []Node

		for _, node := range currentLevel {
			for _, dep := range node.Dependencies() {
				inDegree[dep.NodeID]--
				if inDegree[dep.NodeID] == 0 {
					nextLevel = append(nextLevel, g.Nodes[dep.NodeID])
				}
			}
		}

		currentLevel = nextLevel
	}

	return levels
}
