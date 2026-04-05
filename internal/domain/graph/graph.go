package graph

import (
	"context"
	"sync"
	"time"

	"github.com/Gadzet005/shortcut/pkg/algorithms/topsort"
	"github.com/Gadzet005/shortcut/pkg/containers/slices"
	errors "github.com/Gadzet005/shortcut/pkg/errors"
	"go.uber.org/zap"
)

var _ Graph = graph{}

func NewGraph(
	nodes map[NodeID]Node,
	inputNode NodeID,
	outputNode NodeID,
	timeout time.Duration,
) (graph, error) {
	if _, ok := nodes[inputNode]; !ok {
		return graph{}, errors.Errorf("input node %s not found", inputNode)
	}
	if _, ok := nodes[outputNode]; !ok {
		return graph{}, errors.Errorf("output node %s not found", outputNode)
	}
	return graph{
		nodes:      nodes,
		inputNode:  inputNode,
		outputNode: outputNode,
		timeout:    timeout,
	}, nil
}

type graph struct {
	nodes      map[NodeID]Node
	inputNode  NodeID
	outputNode NodeID
	timeout    time.Duration
}

func (g graph) Run(
	ctx context.Context,
	logger *zap.Logger,
	items map[ItemID]Item,
) (map[ItemID]Item, error) {
	if g.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.timeout)
		defer cancel()
	}
	levelIDs, err := topSort(g, g.inputNode)
	if err != nil {
		return nil, errors.Wrap(err, "top sort by levels")
	}

	results := newGraphResults()
	for itemID, item := range items {
		results.Add(g.inputNode, itemID, item)
	}

	for _, levelNodeIDs := range levelIDs {
		level := make([]Node, 0, len(levelNodeIDs))
		for _, nodeID := range levelNodeIDs {
			level = append(level, g.nodes[nodeID])
		}
		err := visitNodes(ctx, logger, level, results)
		if err != nil {
			return nil, err
		}
	}

	return results.GetAll(g.outputNode), nil
}

// visitNodes посещает ноды паралельно и записывает ответы в results
func visitNodes(
	ctx context.Context,
	logger *zap.Logger,
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

			logger = logger.With(zap.String("node_id", node.ID.String()))

			result, err := visitNode(ctx, logger, node, results)
			if err != nil {
				nodeErrors[i] = err
			}
			tmpResults[i] = result.Items
		}()
	}

	wg.Wait()

	for i, nodeError := range nodeErrors {
		if nodeError != nil {
			return errors.Wrapf(nodeError, "visit node %s", nodes[i].ID)
		}
	}

	for i, result := range tmpResults {
		for name, item := range result {
			results.Add(nodes[i].ID, name, item)
		}
	}
	return nil
}

func visitNode(
	ctx context.Context,
	logger *zap.Logger,
	node Node,
	results graphResults,
) (NodeExecutorResponse, error) {
	items := make(map[ItemID]Item, len(node.Dependencies))
	for _, dep := range node.Dependencies {
		result, ok := results.Get(dep.NodeID, dep.ItemID)
		if !ok {
			return NodeExecutorResponse{}, errors.Error("dependency not found")
		}
		if dep.OverrideItemID != "" {
			items[dep.OverrideItemID] = result
		} else {
			items[dep.ItemID] = result
		}
	}

	resp, err := node.Executor.Run(ctx, logger, NodeExecutorRequest{Items: items})
	if err != nil {
		return NodeExecutorResponse{}, errors.Wrapf(err, "run node %s", node.ID)
	}
	return resp, nil
}

func topSort(graph graph, inputNode NodeID) ([][]NodeID, error) {
	g := map[string][]string{
		inputNode.String(): nil,
	}

	for _, node := range graph.nodes {
		g[node.ID.String()] = nil
	}

	for _, node := range graph.nodes {
		for _, dep := range node.Dependencies {
			n, ok := g[dep.NodeID.String()]
			if !ok {
				return nil, errors.Errorf("dependency not found: %s", dep.NodeID)
			}
			g[dep.NodeID.String()] = append(n, node.ID.String())
		}
	}

	levels, ok := topsort.Sort(g)
	if !ok {
		return nil, errors.Errorf("graph has a cycle")
	}

	convertedLevels := slices.Map(levels, func(level []string) []NodeID {
		return slices.Map(level, func(nodeID string) NodeID {
			return NodeID(nodeID)
		})
	})
	return convertedLevels, nil
}
