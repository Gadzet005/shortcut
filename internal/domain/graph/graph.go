package graph

import (
	"context"
	"time"

	"github.com/Gadzet005/shortcut/pkg/algorithms/topsort"
	errors "github.com/Gadzet005/shortcut/pkg/errors"
	"go.uber.org/zap"
)

const maxParallelism = 100

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
	sem := make(chan struct{}, maxParallelism)

	launch := func(node Node) {
		nodeItems := collectItems(node, results)
		inFlight++
		go func() {
			sem <- struct{}{}
			defer func() { <-sem }()
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

func (g graph) TryRevert(
	ctx context.Context,
	logger *zap.Logger,
	requestID string,
	visitedNodes []NodeID,
) (bool, error) {
	if g.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.timeout)
		defer cancel()
	}

	if err := checkCycle(g, g.inputNode); err != nil {
		return false, err
	}

	visited := make(map[NodeID]struct{}, len(visitedNodes))
	for _, id := range visitedNodes {
		if _, ok := g.nodes[id]; ok {
			visited[id] = struct{}{}
		}
	}
	if len(visited) == 0 {
		return true, nil
	}

	successors := make(map[NodeID][]NodeID)
	for _, node := range g.nodes {
		for _, dep := range node.Dependencies {
			successors[dep.NodeID] = append(successors[dep.NodeID], node.ID)
		}
	}

	remaining := make(map[NodeID]int, len(visited))
	for id := range visited {
		for _, succID := range successors[id] {
			if _, ok := visited[succID]; ok {
				remaining[id]++
			}
		}
	}

	type revertResult struct {
		nodeID NodeID
		ok     bool
		err    error
	}

	completions := make(chan revertResult)
	inFlight := 0
	sem := make(chan struct{}, maxParallelism)

	launch := func(nodeID NodeID) {
		inFlight++
		node := g.nodes[nodeID]
		go func() {
			sem <- struct{}{}
			defer func() { <-sem }()
			ok, err := node.Executor.TryRevert(
				ctx,
				logger.With(zap.String("node_id", nodeID.String())),
				requestID,
			)
			completions <- revertResult{nodeID: nodeID, ok: ok, err: err}
		}()
	}

	for id := range visited {
		if remaining[id] == 0 {
			launch(id)
		}
	}

	allOK := true
	var firstErr error

	for inFlight > 0 {
		res := <-completions
		inFlight--

		if res.err != nil {
			allOK = false
			if firstErr == nil {
				firstErr = errors.Wrapf(res.err, "revert node %s", res.nodeID)
			}
		} else if !res.ok {
			allOK = false
		}

		for _, dep := range g.nodes[res.nodeID].Dependencies {
			if _, ok := visited[dep.NodeID]; !ok {
				continue
			}
			remaining[dep.NodeID]--
			if remaining[dep.NodeID] == 0 {
				launch(dep.NodeID)
			}
		}
	}

	if firstErr != nil {
		return false, firstErr
	}
	return allOK, nil
}

func (g graph) TryFinish(
	ctx context.Context,
	logger *zap.Logger,
	items map[ItemID]Item,
	overrides map[NodeID]string,
	visitedNodes []NodeID,
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
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.timeout)
		defer cancel()
	}

	if err := checkCycle(g, g.inputNode); err != nil {
		return nil, err
	}

	skip := make(map[NodeID]bool, len(visitedNodes))
	for _, id := range visitedNodes {
		if _, ok := g.nodes[id]; ok {
			skip[id] = true
		}
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

	skipQueue := make([]NodeID, 0)
	for nodeID := range g.nodes {
		if skip[nodeID] && remaining[nodeID] == 0 {
			skipQueue = append(skipQueue, nodeID)
		}
	}
	for len(skipQueue) > 0 {
		nodeID := skipQueue[0]
		skipQueue = skipQueue[1:]
		for _, succID := range successors[nodeID] {
			remaining[succID]--
			if remaining[succID] == 0 && skip[succID] {
				skipQueue = append(skipQueue, succID)
			}
		}
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
	launched := make(map[NodeID]bool, len(g.nodes))
	sem := make(chan struct{}, maxParallelism)

	var launch func(node Node)
	var propagate func(nodeID NodeID, items map[ItemID]Item)

	propagate = func(nodeID NodeID, items map[ItemID]Item) {
		for itemID, item := range items {
			results.Add(nodeID, itemID, item)
		}
		for _, succID := range successors[nodeID] {
			remaining[succID]--
			if remaining[succID] != 0 {
				continue
			}
			if skip[succID] {
				propagate(succID, nil)
				continue
			}
			launch(g.nodes[succID])
		}
	}

	launch = func(node Node) {
		if launched[node.ID] {
			return
		}
		launched[node.ID] = true
		nodeItems := collectItems(node, results)
		inFlight++
		go func() {
			sem <- struct{}{}
			defer func() { <-sem }()
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
		if remaining[node.ID] == 0 && !skip[node.ID] {
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

		propagate(res.nodeID, res.items)
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
