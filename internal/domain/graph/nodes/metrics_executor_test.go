package graphnodes

import (
	"sync"
	"testing"
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	mockgraph "github.com/Gadzet005/shortcut/internal/domain/graph/mocks"
	"github.com/Gadzet005/shortcut/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type nodeMetricsCall struct {
	nodeID   graph.NodeID
	nodeType string
	duration time.Duration
	err      error
}

type fakeNodeMetrics struct {
	mu    sync.Mutex
	calls []nodeMetricsCall
}

func (f *fakeNodeMetrics) ObserveRun(nodeID graph.NodeID, nodeType string, duration time.Duration, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, nodeMetricsCall{nodeID, nodeType, duration, err})
}

func (f *fakeNodeMetrics) snapshot() []nodeMetricsCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]nodeMetricsCall, len(f.calls))
	copy(out, f.calls)
	return out
}

func TestMetricsExecutor_Success(t *testing.T) {
	inner := mockgraph.NewNodeExecutor(t)
	inner.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).Return(
		graph.NodeExecutorResponse{Items: map[graph.ItemID]graph.Item{"out": {Data: []byte("ok")}}},
		nil,
	)

	m := &fakeNodeMetrics{}
	exec := NewMetricsExecutor(inner, "node-1", "default", m)

	resp, err := exec.Run(t.Context(), zap.NewNop(), graph.NodeExecutorRequest{})
	require.NoError(t, err)
	require.Equal(t, []byte("ok"), resp.Items["out"].Data)

	calls := m.snapshot()
	require.Equal(t, 1, len(calls))
	require.Equal(t, graph.NodeID("node-1"), calls[0].nodeID)
	require.Equal(t, "default", calls[0].nodeType)
	require.NoError(t, calls[0].err)
	require.GreaterOrEqual(t, calls[0].duration, time.Duration(0))
}

func TestMetricsExecutor_Error(t *testing.T) {
	inner := mockgraph.NewNodeExecutor(t)
	sentinel := errors.Error("boom")
	inner.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).Return(
		graph.NodeExecutorResponse{},
		sentinel,
	)

	m := &fakeNodeMetrics{}
	exec := NewMetricsExecutor(inner, "node-err", "http", m)

	_, err := exec.Run(t.Context(), zap.NewNop(), graph.NodeExecutorRequest{})
	require.ErrorIs(t, err, sentinel)

	calls := m.snapshot()
	require.Equal(t, 1, len(calls))
	require.Equal(t, graph.NodeID("node-err"), calls[0].nodeID)
	require.Equal(t, "http", calls[0].nodeType)
	require.ErrorIs(t, calls[0].err, sentinel)
}

func TestMetricsExecutor_NilMetrics(t *testing.T) {
	inner := mockgraph.NewNodeExecutor(t)
	inner.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).Return(graph.NodeExecutorResponse{}, nil)

	exec := NewMetricsExecutor(inner, "node-2", "default", nil)
	_, err := exec.Run(t.Context(), zap.NewNop(), graph.NodeExecutorRequest{})
	require.NoError(t, err)
}

func TestMetricsExecutor_TryRevertDelegates(t *testing.T) {
	inner := mockgraph.NewNodeExecutor(t)
	inner.EXPECT().TryRevert(mock.Anything, mock.Anything, "req-1").Return(true, nil)

	exec := NewMetricsExecutor(inner, "node-3", "default", &fakeNodeMetrics{})
	ok, err := exec.TryRevert(t.Context(), zap.NewNop(), "req-1")
	require.NoError(t, err)
	require.True(t, ok)
}
