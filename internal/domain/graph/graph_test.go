package graph_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	mockgraph "github.com/Gadzet005/shortcut/internal/domain/graph/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGraphRun_Linear(t *testing.T) {
	inputExec := mockgraph.NewNodeExecutor(t)
	middleExec := mockgraph.NewNodeExecutor(t)
	outputExec := mockgraph.NewNodeExecutor(t)

	itemA := itemOf("a")
	itemB := itemOf("b")
	itemC := itemOf("c")

	inputExec.EXPECT().
		Run(mock.Anything, mock.Anything, matchItems(map[graph.ItemID]graph.Item{})).
		Return(respOf(nil), nil)

	middleExec.EXPECT().
		Run(mock.Anything, mock.Anything, matchItems(map[graph.ItemID]graph.Item{"x": itemA})).
		Return(respOf(map[graph.ItemID]graph.Item{"y": itemB}), nil)

	outputExec.EXPECT().
		Run(mock.Anything, mock.Anything, matchItems(map[graph.ItemID]graph.Item{"y": itemB})).
		Return(respOf(map[graph.ItemID]graph.Item{"z": itemC}), nil)

	nodes := map[graph.NodeID]graph.Node{
		"input": {ID: "input", Executor: inputExec},
		"middle": {
			ID:           "middle",
			Dependencies: []graph.Dependency{{NodeID: "input", ItemID: "x"}},
			Executor:     middleExec,
		},
		"output": {
			ID:           "output",
			Dependencies: []graph.Dependency{{NodeID: "middle", ItemID: "y"}},
			Executor:     outputExec,
		},
	}
	g, err := graph.NewGraph(nodes, "input", "output", 0)
	require.NoError(t, err)

	result, err := g.Run(t.Context(), zap.NewNop(), map[graph.ItemID]graph.Item{"x": itemA}, nil)
	require.NoError(t, err)
	require.Equal(t, map[graph.ItemID]graph.Item{"z": itemC}, result)
}

func TestGraphRun_Parallel(t *testing.T) {
	inputExec := mockgraph.NewNodeExecutor(t)
	execA := mockgraph.NewNodeExecutor(t)
	execB := mockgraph.NewNodeExecutor(t)
	outputExec := mockgraph.NewNodeExecutor(t)

	itemIn := itemOf("in")
	itemA := itemOf("a")
	itemB := itemOf("b")
	itemOut := itemOf("out")

	inputExec.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).
		Return(respOf(nil), nil)

	var barrier sync.WaitGroup
	barrier.Add(2)
	execA.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, _ *zap.Logger, _ graph.NodeExecutorRequest) (graph.NodeExecutorResponse, error) {
			barrier.Done()
			barrier.Wait()
			return respOf(map[graph.ItemID]graph.Item{"a": itemA}), nil
		})
	execB.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, _ *zap.Logger, _ graph.NodeExecutorRequest) (graph.NodeExecutorResponse, error) {
			barrier.Done()
			barrier.Wait()
			return respOf(map[graph.ItemID]graph.Item{"b": itemB}), nil
		})

	outputExec.EXPECT().
		Run(mock.Anything, mock.Anything, matchItems(map[graph.ItemID]graph.Item{"a": itemA, "b": itemB})).
		Return(respOf(map[graph.ItemID]graph.Item{"out": itemOut}), nil)

	nodes := map[graph.NodeID]graph.Node{
		"input": {ID: "input", Executor: inputExec},
		"nodeA": {
			ID:           "nodeA",
			Dependencies: []graph.Dependency{{NodeID: "input", ItemID: "in"}},
			Executor:     execA,
		},
		"nodeB": {
			ID:           "nodeB",
			Dependencies: []graph.Dependency{{NodeID: "input", ItemID: "in"}},
			Executor:     execB,
		},
		"output": {
			ID: "output",
			Dependencies: []graph.Dependency{
				{NodeID: "nodeA", ItemID: "a"},
				{NodeID: "nodeB", ItemID: "b"},
			},
			Executor: outputExec,
		},
	}
	g, err := graph.NewGraph(nodes, "input", "output", 0)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := g.Run(ctx, zap.NewNop(), map[graph.ItemID]graph.Item{"in": itemIn}, nil)
	require.NoError(t, err)
	require.Equal(t, map[graph.ItemID]graph.Item{"out": itemOut}, result)
}

func TestGraphRun_NodeError(t *testing.T) {
	inputExec := mockgraph.NewNodeExecutor(t)
	failExec := mockgraph.NewNodeExecutor(t)

	inputExec.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).
		Return(respOf(nil), nil)
	failExec.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).
		Return(graph.NodeExecutorResponse{}, errors.New("something went wrong"))

	nodes := map[graph.NodeID]graph.Node{
		"input": {ID: "input", Executor: inputExec},
		"fail": {
			ID:           "fail",
			Dependencies: []graph.Dependency{{NodeID: "input", ItemID: "x"}},
			Executor:     failExec,
		},
	}
	g, err := graph.NewGraph(nodes, "input", "fail", 0)
	require.NoError(t, err)

	_, err = g.Run(t.Context(), zap.NewNop(), map[graph.ItemID]graph.Item{"x": itemOf("x")}, nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "something went wrong")
}

func TestGraphRun_Cycle(t *testing.T) {
	execA := mockgraph.NewNodeExecutor(t)
	execB := mockgraph.NewNodeExecutor(t)

	nodes := map[graph.NodeID]graph.Node{
		"A": {
			ID:           "A",
			Dependencies: []graph.Dependency{{NodeID: "B", ItemID: "x"}},
			Executor:     execA,
		},
		"B": {
			ID:           "B",
			Dependencies: []graph.Dependency{{NodeID: "A", ItemID: "y"}},
			Executor:     execB,
		},
	}
	g, err := graph.NewGraph(nodes, "A", "B", 0)
	require.NoError(t, err)

	_, err = g.Run(t.Context(), zap.NewNop(), nil, nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "cycle")
}

func TestGraphRun_UnknownOverrideNode(t *testing.T) {
	exec := mockgraph.NewNodeExecutor(t)
	nodes := map[graph.NodeID]graph.Node{
		"input": {ID: "input", Executor: exec},
	}
	g, err := graph.NewGraph(nodes, "input", "input", 0)
	require.NoError(t, err)

	_, err = g.Run(t.Context(), zap.NewNop(), nil, map[graph.NodeID]string{"nonexistent": "host:1234"})
	require.Error(t, err)

	var nodeErr *graph.NodeError
	require.ErrorAs(t, err, &nodeErr)
	require.Equal(t, graph.ErrCodeBadRequest, nodeErr.Code)
}

func TestGraphRun_EndpointOverride(t *testing.T) {
	inputExec := mockgraph.NewNodeExecutor(t)
	targetExec := mockgraph.NewNodeExecutor(t)

	overrideAddr := "new-host:9090"

	inputExec.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).
		Return(respOf(nil), nil)
	targetExec.EXPECT().
		Run(mock.Anything, mock.Anything, mock.MatchedBy(func(req graph.NodeExecutorRequest) bool {
			return req.EndpointOverride != nil && *req.EndpointOverride == overrideAddr
		})).
		Return(respOf(map[graph.ItemID]graph.Item{"y": itemOf("y")}), nil)

	nodes := map[graph.NodeID]graph.Node{
		"input": {ID: "input", Executor: inputExec},
		"target": {
			ID:           "target",
			Dependencies: []graph.Dependency{{NodeID: "input", ItemID: "x"}},
			Executor:     targetExec,
		},
	}
	g, err := graph.NewGraph(nodes, "input", "target", 0)
	require.NoError(t, err)

	_, err = g.Run(
		t.Context(), zap.NewNop(),
		map[graph.ItemID]graph.Item{"x": itemOf("x")},
		map[graph.NodeID]string{"target": overrideAddr},
	)
	require.NoError(t, err)
}

// TestGraphRun_Timeout verifies that the graph timeout cancels a hanging node.
func TestGraphRun_Timeout(t *testing.T) {
	inputExec := mockgraph.NewNodeExecutor(t)
	slowExec := mockgraph.NewNodeExecutor(t)

	inputExec.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).
		Return(respOf(nil), nil)
	slowExec.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, _ *zap.Logger, _ graph.NodeExecutorRequest) (graph.NodeExecutorResponse, error) {
			<-ctx.Done()
			return graph.NodeExecutorResponse{}, ctx.Err()
		})

	nodes := map[graph.NodeID]graph.Node{
		"input": {ID: "input", Executor: inputExec},
		"slow": {
			ID:           "slow",
			Dependencies: []graph.Dependency{{NodeID: "input", ItemID: "x"}},
			Executor:     slowExec,
		},
	}
	g, err := graph.NewGraph(nodes, "input", "slow", 50*time.Millisecond)
	require.NoError(t, err)

	_, err = g.Run(t.Context(), zap.NewNop(), map[graph.ItemID]graph.Item{"x": itemOf("x")}, nil)
	require.Error(t, err)
}

func itemOf(s string) graph.Item {
	return graph.Item{Data: []byte(s)}
}

func respOf(items map[graph.ItemID]graph.Item) graph.NodeExecutorResponse {
	return graph.NodeExecutorResponse{Items: items}
}

func matchItems(expected map[graph.ItemID]graph.Item) any {
	return mock.MatchedBy(func(req graph.NodeExecutorRequest) bool {
		if len(req.Items) != len(expected) {
			return false
		}
		for k, v := range expected {
			got, ok := req.Items[k]
			if !ok || string(got.Data) != string(v.Data) {
				return false
			}
		}
		return true
	})
}
