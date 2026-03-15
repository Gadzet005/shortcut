package graph

import (
	"context"
	"sync"

	errors "github.com/Gadzet005/shortcut/pkg/errors"
	"go.uber.org/zap"
)

type ID string

func (i ID) String() string {
	return string(i)
}

type Graph struct {
	ID    ID
	Nodes map[NodeID]Node
	FailureStrategy FailureStrategy 
}

func (g Graph) Run(
	ctx context.Context,
	logger *zap.Logger,
	req RunNodeRequest,
) (RunNodeResponse, error) {
	results := newGraphResults()
	for id, item := range req.Items {
		results.Add(InputNodeID, id, item)
	}

	levelIDs, err := TopSort(g)

	if err != nil {
		return RunNodeResponse{}, errors.Wrap(err, "top sort by levels")
	}
	levelIDs = levelIDs[1:] // входную вершину не посещаем

	for _, levelNodeIDs := range levelIDs {
		level := make([]Node, 0, len(levelNodeIDs))
		for _, nodeID := range levelNodeIDs {
			node, exists := g.Nodes[nodeID]
			if !exists {
				return RunNodeResponse{}, errors.Errorf("node %s not found", nodeID)
			}
			level = append(level, node)
		}

		err := visitNodes(ctx, logger, req, level, results)
		if err != nil {
			return RunNodeResponse{}, err
		}
	}

	lastNodeID := levelIDs[len(levelIDs)-1][0]
	lastNodeData, ok := results.GetAny(lastNodeID)
	if !ok {
		return RunNodeResponse{}, errors.Error("failed to get last node data")
	}
	return RunNodeResponse{Items: map[ItemID]Item{
		DefaultItemID: lastNodeData,
	}}, nil
}

// visitNodes посещает ноды паралельно и записывает ответы в results
func visitNodes(
	ctx context.Context,
	logger *zap.Logger,
	req RunNodeRequest,
	nodes []Node,
	results graphResults,
) error {
	tmpResults := make([]map[ItemID]Item, len(nodes))
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
			return errors.Wrapf(nodeError, "visit node %s", nodes[i].ID())
		}
	}

	for i, result := range tmpResults {
		for name, item := range result {
			results.Add(nodes[i].ID(), name, item)
		}
	}
	return nil
}

func visitNode(
	ctx context.Context,
	logger *zap.Logger,
	req RunNodeRequest,
	node Node,
	results graphResults,
) (map[ItemID]Item, error) {
	items := make(map[ItemID]Item, len(node.Dependencies()))
	for _, dep := range node.Dependencies() {
		result, ok := results.Get(dep.NodeID, dep.ItemID)
		if !ok {
			return nil, errors.Error("failed to get dependency")
		}
		items[dep.OverridenItemID] = result
	}

	resp, err := node.Run(ctx, logger, RunNodeRequest{
		Client: req.Client,
		Items:  items,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "run node %s", node.ID())
	}
	return resp.Items, nil
}
