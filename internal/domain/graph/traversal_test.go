package graph_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	mockgraph "github.com/Gadzet005/shortcut/internal/domain/graph/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGraphTryRevert_Linear(t *testing.T) {
	inputExec := mockgraph.NewNodeExecutor(t)
	middleExec := mockgraph.NewNodeExecutor(t)
	outputExec := mockgraph.NewNodeExecutor(t)

	var order []string
	var mu sync.Mutex
	record := func(id string) {
		mu.Lock()
		defer mu.Unlock()
		order = append(order, id)
	}

	inputExec.EXPECT().TryRevert(mock.Anything, mock.Anything, "req").
		RunAndReturn(func(_ context.Context, _ *zap.Logger, _ string) (bool, error) {
			record("input")
			return true, nil
		})
	middleExec.EXPECT().TryRevert(mock.Anything, mock.Anything, "req").
		RunAndReturn(func(_ context.Context, _ *zap.Logger, _ string) (bool, error) {
			record("middle")
			return true, nil
		})
	outputExec.EXPECT().TryRevert(mock.Anything, mock.Anything, "req").
		RunAndReturn(func(_ context.Context, _ *zap.Logger, _ string) (bool, error) {
			record("output")
			return true, nil
		})

	nodes := map[graph.NodeID]graph.Node{
		"input":  {ID: "input", Executor: inputExec},
		"middle": {ID: "middle", Dependencies: []graph.Dependency{{NodeID: "input", ItemID: "x"}}, Executor: middleExec},
		"output": {ID: "output", Dependencies: []graph.Dependency{{NodeID: "middle", ItemID: "y"}}, Executor: outputExec},
	}
	g, err := graph.NewGraph(nodes, "input", "output", 0)
	require.NoError(t, err)

	ok, err := g.TryRevert(t.Context(), zap.NewNop(), "req", []graph.NodeID{"input", "middle", "output"})
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, []string{"output", "middle", "input"}, order)
}

func TestGraphTryRevert_OnlyVisitedNodes(t *testing.T) {
	inputExec := mockgraph.NewNodeExecutor(t)
	middleExec := mockgraph.NewNodeExecutor(t)
	outputExec := mockgraph.NewNodeExecutor(t)

	inputExec.EXPECT().TryRevert(mock.Anything, mock.Anything, "req").Return(true, nil)
	middleExec.EXPECT().TryRevert(mock.Anything, mock.Anything, "req").Return(true, nil)

	nodes := map[graph.NodeID]graph.Node{
		"input":  {ID: "input", Executor: inputExec},
		"middle": {ID: "middle", Dependencies: []graph.Dependency{{NodeID: "input", ItemID: "x"}}, Executor: middleExec},
		"output": {ID: "output", Dependencies: []graph.Dependency{{NodeID: "middle", ItemID: "y"}}, Executor: outputExec},
	}
	g, err := graph.NewGraph(nodes, "input", "output", 0)
	require.NoError(t, err)

	ok, err := g.TryRevert(t.Context(), zap.NewNop(), "req", []graph.NodeID{"input", "middle"})
	require.NoError(t, err)
	require.True(t, ok)
}

func TestGraphTryRevert_ParallelLeaves(t *testing.T) {
	inputExec := mockgraph.NewNodeExecutor(t)
	leafA := mockgraph.NewNodeExecutor(t)
	leafB := mockgraph.NewNodeExecutor(t)

	var barrier sync.WaitGroup
	barrier.Add(2)

	leafA.EXPECT().TryRevert(mock.Anything, mock.Anything, "req").
		RunAndReturn(func(_ context.Context, _ *zap.Logger, _ string) (bool, error) {
			barrier.Done()
			barrier.Wait()
			return true, nil
		})
	leafB.EXPECT().TryRevert(mock.Anything, mock.Anything, "req").
		RunAndReturn(func(_ context.Context, _ *zap.Logger, _ string) (bool, error) {
			barrier.Done()
			barrier.Wait()
			return true, nil
		})
	inputExec.EXPECT().TryRevert(mock.Anything, mock.Anything, "req").Return(true, nil)

	nodes := map[graph.NodeID]graph.Node{
		"input": {ID: "input", Executor: inputExec},
		"leafA": {ID: "leafA", Dependencies: []graph.Dependency{{NodeID: "input", ItemID: "x"}}, Executor: leafA},
		"leafB": {ID: "leafB", Dependencies: []graph.Dependency{{NodeID: "input", ItemID: "x"}}, Executor: leafB},
	}
	g, err := graph.NewGraph(nodes, "input", "leafA", 0)
	require.NoError(t, err)

	ok, err := g.TryRevert(t.Context(), zap.NewNop(), "req", []graph.NodeID{"input", "leafA", "leafB"})
	require.NoError(t, err)
	require.True(t, ok)
}

func TestGraphTryRevert_ReturnsFalseOnNodeFailure(t *testing.T) {
	inputExec := mockgraph.NewNodeExecutor(t)
	leafExec := mockgraph.NewNodeExecutor(t)

	leafExec.EXPECT().TryRevert(mock.Anything, mock.Anything, "req").Return(false, nil)
	inputExec.EXPECT().TryRevert(mock.Anything, mock.Anything, "req").Return(true, nil)

	nodes := map[graph.NodeID]graph.Node{
		"input": {ID: "input", Executor: inputExec},
		"leaf":  {ID: "leaf", Dependencies: []graph.Dependency{{NodeID: "input", ItemID: "x"}}, Executor: leafExec},
	}
	g, err := graph.NewGraph(nodes, "input", "leaf", 0)
	require.NoError(t, err)

	ok, err := g.TryRevert(t.Context(), zap.NewNop(), "req", []graph.NodeID{"input", "leaf"})
	require.NoError(t, err)
	require.False(t, ok)
}

func TestGraphTryRevert_PropagatesError(t *testing.T) {
	inputExec := mockgraph.NewNodeExecutor(t)
	leafExec := mockgraph.NewNodeExecutor(t)

	leafExec.EXPECT().TryRevert(mock.Anything, mock.Anything, "req").
		Return(false, errors.New("boom"))
	inputExec.EXPECT().TryRevert(mock.Anything, mock.Anything, "req").Return(true, nil).Maybe()

	nodes := map[graph.NodeID]graph.Node{
		"input": {ID: "input", Executor: inputExec},
		"leaf":  {ID: "leaf", Dependencies: []graph.Dependency{{NodeID: "input", ItemID: "x"}}, Executor: leafExec},
	}
	g, err := graph.NewGraph(nodes, "input", "leaf", 0)
	require.NoError(t, err)

	ok, err := g.TryRevert(t.Context(), zap.NewNop(), "req", []graph.NodeID{"input", "leaf"})
	require.Error(t, err)
	require.False(t, ok)
}

func TestGraphTryRevert_EmptyVisitedNoOp(t *testing.T) {
	exec := mockgraph.NewNodeExecutor(t)
	nodes := map[graph.NodeID]graph.Node{
		"input": {ID: "input", Executor: exec},
	}
	g, err := graph.NewGraph(nodes, "input", "input", 0)
	require.NoError(t, err)

	ok, err := g.TryRevert(t.Context(), zap.NewNop(), "req", nil)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestGraphTryFinish_SkipsVisitedNodes(t *testing.T) {
	inputExec := mockgraph.NewNodeExecutor(t)
	middleExec := mockgraph.NewNodeExecutor(t)
	outputExec := mockgraph.NewNodeExecutor(t)

	var middleCalls atomic.Int32
	middleExec.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, _ *zap.Logger, _ graph.NodeExecutorRequest) (graph.NodeExecutorResponse, error) {
			middleCalls.Add(1)
			return graph.NodeExecutorResponse{Items: map[graph.ItemID]graph.Item{"y": itemOf("y")}}, nil
		})
	outputExec.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).
		Return(graph.NodeExecutorResponse{Items: map[graph.ItemID]graph.Item{"z": itemOf("z")}}, nil)

	nodes := map[graph.NodeID]graph.Node{
		"input":  {ID: "input", Executor: inputExec},
		"middle": {ID: "middle", Dependencies: []graph.Dependency{{NodeID: "input", ItemID: "x"}}, Executor: middleExec},
		"output": {ID: "output", Dependencies: []graph.Dependency{{NodeID: "middle", ItemID: "y"}}, Executor: outputExec},
	}
	g, err := graph.NewGraph(nodes, "input", "output", 0)
	require.NoError(t, err)

	result, err := g.TryFinish(
		t.Context(), zap.NewNop(),
		map[graph.ItemID]graph.Item{"x": itemOf("x")},
		nil,
		[]graph.NodeID{"input"},
	)
	require.NoError(t, err)
	require.Equal(t, int32(1), middleCalls.Load())
	require.Equal(t, map[graph.ItemID]graph.Item{"z": itemOf("z")}, result)
}

func TestGraphTryFinish_AllVisitedNoExecution(t *testing.T) {
	inputExec := mockgraph.NewNodeExecutor(t)
	outputExec := mockgraph.NewNodeExecutor(t)

	nodes := map[graph.NodeID]graph.Node{
		"input":  {ID: "input", Executor: inputExec},
		"output": {ID: "output", Dependencies: []graph.Dependency{{NodeID: "input", ItemID: "x"}}, Executor: outputExec},
	}
	g, err := graph.NewGraph(nodes, "input", "output", 0)
	require.NoError(t, err)

	result, err := g.TryFinish(
		t.Context(), zap.NewNop(),
		map[graph.ItemID]graph.Item{"x": itemOf("x")},
		nil,
		[]graph.NodeID{"input", "output"},
	)
	require.NoError(t, err)
	require.Empty(t, result)
}

func TestGraphTryFinish_FailedNodePropagatesError(t *testing.T) {
	inputExec := mockgraph.NewNodeExecutor(t)
	failExec := mockgraph.NewNodeExecutor(t)

	failExec.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).
		Return(graph.NodeExecutorResponse{}, errors.New("still broken"))

	nodes := map[graph.NodeID]graph.Node{
		"input": {ID: "input", Executor: inputExec},
		"fail":  {ID: "fail", Dependencies: []graph.Dependency{{NodeID: "input", ItemID: "x"}}, Executor: failExec},
	}
	g, err := graph.NewGraph(nodes, "input", "fail", 0)
	require.NoError(t, err)

	_, err = g.TryFinish(
		t.Context(), zap.NewNop(),
		map[graph.ItemID]graph.Item{"x": itemOf("x")},
		nil,
		[]graph.NodeID{"input"},
	)
	require.Error(t, err)
	require.ErrorContains(t, err, "still broken")
}

func TestGraphTryFinish_UnknownOverride(t *testing.T) {
	exec := mockgraph.NewNodeExecutor(t)
	nodes := map[graph.NodeID]graph.Node{
		"input": {ID: "input", Executor: exec},
	}
	g, err := graph.NewGraph(nodes, "input", "input", 0)
	require.NoError(t, err)

	_, err = g.TryFinish(t.Context(), zap.NewNop(), nil, map[graph.NodeID]string{"missing": "host:1"}, nil)
	require.Error(t, err)
	var nerr *graph.NodeError
	require.ErrorAs(t, err, &nerr)
}
