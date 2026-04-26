package graph

import (
	"context"
	"time"

	"github.com/Gadzet005/shortcut/pkg/algorithms/topsort"
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
	overrides map[NodeID]string,
) (map[ItemID]Item, error) {
	for nodeID := range overrides {
		if _, ok := g.nodes[nodeID]; !ok {
			return nil, &NodeError{
				Code:    ErrCodeBadRequest,
				Payload: map[string]any{"error": "node " + nodeID.String() + " not found"},
			}
		}
	}

	if g.timeout > 0 {
		var timeoutCancel context.CancelFunc
		ctx, timeoutCancel = context.WithTimeout(ctx, g.timeout)
		defer timeoutCancel()
	}

	if err := checkCycle(g, g.inputNode); err != nil {
		return nil, err
	}

	remaining := make(map[NodeID]int, len(g.nodes))
	successors := make(map[NodeID][]NodeID)
	for nodeID := range g.nodes {
		remaining[nodeID] = 0
	}
	for _, node := range g.nodes {
		for _, dep := range node.Dependencies {
			successors[dep.NodeID] = append(successors[dep.NodeID], node.ID)
			remaining[node.ID]++
		}
	}

	results := newGraphResults()
	for itemID, item := range items {
		results.Add(g.inputNode, itemID, item)
	}

	type nodeResult struct {
		nodeID NodeID
		items  map[ItemID]Item
		err    error
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	completions := make(chan nodeResult)
	inFlight := 0

	launch := func(node Node) {
		nodeItems := collectItems(node, results)
		inFlight++
		go func() {
			req := NodeExecutorRequest{Items: nodeItems}
			if override, ok := overrides[node.ID]; ok {
				req.EndpointOverride = &override
			}
			resp, err := node.Executor.Run(
				ctx,
				logger.With(zap.String("node_id", node.ID.String())),
				req,
			)
			if err != nil {
				completions <- nodeResult{nodeID: node.ID, err: errors.Wrapf(err, "run node %s", node.ID)}
				return
			}
			completions <- nodeResult{nodeID: node.ID, items: resp.Items}
		}()
	}

	for _, node := range g.nodes {
		if remaining[node.ID] == 0 {
			launch(node)
		}
	}

	var firstErr error
	for inFlight > 0 {
		res := <-completions
		inFlight--

		if res.err != nil {
			if firstErr == nil {
				firstErr = res.err
				cancel()
			}
			continue
		}

		if firstErr != nil {
			continue
		}

		for itemID, item := range res.items {
			results.Add(res.nodeID, itemID, item)
		}

		for _, succID := range successors[res.nodeID] {
			remaining[succID]--
			if remaining[succID] == 0 {
				launch(g.nodes[succID])
			}
		}
	}

	if firstErr != nil {
		return nil, firstErr
	}

	return results.GetAll(g.outputNode), nil
}

func collectItems(node Node, results graphResults) map[ItemID]Item {
	items := make(map[ItemID]Item, len(node.Dependencies))
	for _, dep := range node.Dependencies {
		result, ok := results.Get(dep.NodeID, dep.ItemID)
		if !ok {
			continue
		}
		if dep.OverrideItemID != "" {
			items[dep.OverrideItemID] = result
		} else {
			items[dep.ItemID] = result
		}
	}
	return items
}

func checkCycle(g graph, inputNode NodeID) error {
	adj := map[string][]string{
		inputNode.String(): nil,
	}
	for _, node := range g.nodes {
		adj[node.ID.String()] = nil
	}
	for _, node := range g.nodes {
		for _, dep := range node.Dependencies {
			n, ok := adj[dep.NodeID.String()]
			if !ok {
				return errors.Errorf("dependency not found: %s", dep.NodeID)
			}
			adj[dep.NodeID.String()] = append(n, node.ID.String())
		}
	}
	_, ok := topsort.Sort(adj)
	if !ok {
		return errors.Errorf("graph has a cycle")
	}
	return nil
}
