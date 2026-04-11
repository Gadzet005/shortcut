package trace

import (
	"context"
	"testing"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	mockgraph "github.com/Gadzet005/shortcut/internal/domain/graph/mocks"
	"github.com/Gadzet005/shortcut/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestTracingExecutor_NoCollector(t *testing.T) {
	inner := mockgraph.NewNodeExecutor(t)
	inner.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).Return(
		graph.NodeExecutorResponse{
			Items: map[graph.ItemID]graph.Item{"out": {Data: []byte("data")}},
		},
		nil,
	)
	exec := NewTracingExecutor(inner, "node-1", "default", nil)

	resp, err := exec.Run(t.Context(), zap.NewNop(), graph.NodeExecutorRequest{})
	require.NoError(t, err)
	require.Equal(t, "data", string(resp.Items["out"].Data))
}

func TestTracingExecutor_WithCollector(t *testing.T) {
	inner := mockgraph.NewNodeExecutor(t)
	inner.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).Return(
		graph.NodeExecutorResponse{
			Items: map[graph.ItemID]graph.Item{"out": {Data: []byte("data")}},
			Meta:  map[string]any{"status_code": 200, "retry_count": 0},
		},
		nil,
	)
	exec := NewTracingExecutor(inner, "node-1", "default", nil)

	collector := NewCollector("req-1")
	ctx := WithCollector(t.Context(), collector)

	resp, err := exec.Run(ctx, zap.NewNop(), graph.NodeExecutorRequest{})
	require.NoError(t, err)
	require.Equal(t, "data", string(resp.Items["out"].Data))

	traces := collector.NodeTraces()
	require.Equal(t, 1, len(traces))
	nt := traces[0]
	require.Equal(t, "node-1", nt.NodeID)
	require.Equal(t, 200, nt.StatusCode)
	require.Equal(t, 0, nt.RetryCount)
	require.Equal(t, "", nt.Error)
	require.GreaterOrEqual(t, nt.DurationMs, int64(0))
}

func TestTracingExecutor_WithError(t *testing.T) {
	inner := mockgraph.NewNodeExecutor(t)
	inner.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).Return(
		graph.NodeExecutorResponse{},
		&graph.NodeError{Code: graph.ErrCodeBadRequest, Payload: map[string]any{"msg": "bad"}},
	)
	exec := NewTracingExecutor(inner, "node-err", "default", nil)

	collector := NewCollector("req-2")
	ctx := WithCollector(t.Context(), collector)

	_, err := exec.Run(ctx, zap.NewNop(), graph.NodeExecutorRequest{})
	if err == nil {
		t.Fatal("expected error")
	}

	traces := collector.NodeTraces()
	require.Equal(t, 1, len(traces))
	nt := traces[0]
	require.Equal(t, 400, nt.StatusCode)
	require.NotEmpty(t, nt.Error)
}

func TestTracingExecutor_WrappedError(t *testing.T) {
	nodeErr := &graph.NodeError{Code: graph.ErrCodeNotFound, Payload: map[string]any{}}
	inner := mockgraph.NewNodeExecutor(t)
	inner.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).Return(
		graph.NodeExecutorResponse{},
		errors.Wrap(nodeErr, "wrapped"),
	)
	exec := NewTracingExecutor(inner, "node-wrapped", "default", nil)

	collector := NewCollector("req-3")
	ctx := WithCollector(t.Context(), collector)

	_, err := exec.Run(ctx, zap.NewNop(), graph.NodeExecutorRequest{})
	require.Error(t, err)

	traces := collector.NodeTraces()
	nt := traces[0]
	require.Equal(t, 404, nt.StatusCode)
}

func TestTracingExecutor_MetaOverridesErrorCode(t *testing.T) {
	inner := mockgraph.NewNodeExecutor(t)
	inner.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).Return(
		graph.NodeExecutorResponse{
			Meta: map[string]any{"status_code": 503},
		},
		&graph.NodeError{Code: graph.ErrCodeInternal},
	)
	exec := NewTracingExecutor(inner, "node-meta", "default", nil)

	collector := NewCollector("req-4")
	ctx := WithCollector(context.Background(), collector)

	exec.Run(ctx, zap.NewNop(), graph.NodeExecutorRequest{})

	traces := collector.NodeTraces()
	nt := traces[0]
	require.Equal(t, 503, nt.StatusCode)
}
