package trace

import (
	"context"
	"testing"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	"github.com/Gadzet005/shortcut/pkg/errors"
	"go.uber.org/zap"
)

type mockExecutor struct {
	resp graph.NodeExecutorResponse
	err  error
}

func (m mockExecutor) Run(
	_ context.Context,
	_ *zap.Logger,
	_ graph.NodeExecutorRequest,
) (graph.NodeExecutorResponse, error) {
	return m.resp, m.err
}

func TestTracingExecutor_NoCollector(t *testing.T) {
	inner := mockExecutor{
		resp: graph.NodeExecutorResponse{
			Items: map[graph.ItemID]graph.Item{"out": {Data: []byte("data")}},
		},
	}
	exec := NewTracingExecutor(inner, "node-1")

	resp, err := exec.Run(context.Background(), zap.NewNop(), graph.NodeExecutorRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := resp.Items["out"]; !ok {
		t.Error("expected 'out' item in response")
	}
}

func TestTracingExecutor_WithCollector(t *testing.T) {
	inner := mockExecutor{
		resp: graph.NodeExecutorResponse{
			Items: map[graph.ItemID]graph.Item{"out": {Data: []byte("data")}},
			Meta:  map[string]any{"status_code": 200, "retry_count": 0},
		},
	}
	exec := NewTracingExecutor(inner, "node-1")

	collector := NewCollector("req-1")
	ctx := WithCollector(context.Background(), collector)

	resp, err := exec.Run(ctx, zap.NewNop(), graph.NodeExecutorRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := resp.Items["out"]; !ok {
		t.Error("expected 'out' item in response")
	}

	traces := collector.NodeTraces()
	if len(traces) != 1 {
		t.Fatalf("expected 1 trace, got %d", len(traces))
	}
	nt := traces[0]
	if nt.NodeID != "node-1" {
		t.Errorf("expected node ID 'node-1', got '%s'", nt.NodeID)
	}
	if nt.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", nt.StatusCode)
	}
	if nt.RetryCount != 0 {
		t.Errorf("expected retry count 0, got %d", nt.RetryCount)
	}
	if nt.Error != "" {
		t.Errorf("expected no error, got '%s'", nt.Error)
	}
	if nt.DurationMs < 0 {
		t.Errorf("expected non-negative duration, got %d", nt.DurationMs)
	}
}

func TestTracingExecutor_WithError(t *testing.T) {
	inner := mockExecutor{
		err: &graph.NodeError{Code: graph.ErrCodeBadRequest, Payload: map[string]any{"msg": "bad"}},
	}
	exec := NewTracingExecutor(inner, "node-err")

	collector := NewCollector("req-2")
	ctx := WithCollector(context.Background(), collector)

	_, err := exec.Run(ctx, zap.NewNop(), graph.NodeExecutorRequest{})
	if err == nil {
		t.Fatal("expected error")
	}

	traces := collector.NodeTraces()
	if len(traces) != 1 {
		t.Fatalf("expected 1 trace, got %d", len(traces))
	}
	nt := traces[0]
	if nt.StatusCode != 400 {
		t.Errorf("expected status code 400, got %d", nt.StatusCode)
	}
	if nt.Error == "" {
		t.Error("expected error message in trace")
	}
}

func TestTracingExecutor_WrappedError(t *testing.T) {
	nodeErr := &graph.NodeError{Code: graph.ErrCodeNotFound, Payload: map[string]any{}}
	inner := mockExecutor{
		err: errors.Wrap(nodeErr, "wrapped"),
	}
	exec := NewTracingExecutor(inner, "node-wrapped")

	collector := NewCollector("req-3")
	ctx := WithCollector(context.Background(), collector)

	_, err := exec.Run(ctx, zap.NewNop(), graph.NodeExecutorRequest{})
	if err == nil {
		t.Fatal("expected error")
	}

	traces := collector.NodeTraces()
	nt := traces[0]
	if nt.StatusCode != 404 {
		t.Errorf("expected status code 404, got %d", nt.StatusCode)
	}
}

func TestTracingExecutor_MetaOverridesErrorCode(t *testing.T) {
	inner := mockExecutor{
		resp: graph.NodeExecutorResponse{
			Meta: map[string]any{"status_code": 503},
		},
		err: &graph.NodeError{Code: graph.ErrCodeInternal},
	}
	exec := NewTracingExecutor(inner, "node-meta")

	collector := NewCollector("req-4")
	ctx := WithCollector(context.Background(), collector)

	exec.Run(ctx, zap.NewNop(), graph.NodeExecutorRequest{})

	traces := collector.NodeTraces()
	nt := traces[0]
	// Meta status_code should take precedence
	if nt.StatusCode != 503 {
		t.Errorf("expected status code 503 from Meta, got %d", nt.StatusCode)
	}
}
