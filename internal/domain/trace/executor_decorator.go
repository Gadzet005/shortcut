package trace

import (
	"context"
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	"github.com/Gadzet005/shortcut/pkg/errors"
	"go.uber.org/zap"
)

var _ graph.NodeExecutor = tracingExecutor{}

func NewTracingExecutor(inner graph.NodeExecutor, nodeID graph.NodeID) graph.NodeExecutor {
	return tracingExecutor{inner: inner, nodeID: nodeID}
}

type tracingExecutor struct {
	inner  graph.NodeExecutor
	nodeID graph.NodeID
}

func (e tracingExecutor) Run(
	ctx context.Context,
	logger *zap.Logger,
	req graph.NodeExecutorRequest,
) (graph.NodeExecutorResponse, error) {
	collector, ok := GetCollector(ctx)
	if !ok {
		return e.inner.Run(ctx, logger, req)
	}

	start := time.Now()
	resp, err := e.inner.Run(ctx, logger, req)
	finished := time.Now()

	nt := NodeTrace{
		NodeID:     e.nodeID.String(),
		StartedAt:  start,
		FinishedAt: finished,
		DurationMs: finished.Sub(start).Milliseconds(),
	}

	if resp.Meta != nil {
		if sc, ok := resp.Meta["status_code"].(int); ok {
			nt.StatusCode = sc
		}
		if rc, ok := resp.Meta["retry_count"].(int); ok {
			nt.RetryCount = rc
		}
	}

	if err != nil {
		nt.Error = err.Error()
		var nodeErr *graph.NodeError
		if errors.As(err, &nodeErr) && nt.StatusCode == 0 {
			nt.StatusCode = errorCodeToHTTPStatus(nodeErr.Code)
		}
	}

	collector.Add(nt)
	return resp, err
}

func errorCodeToHTTPStatus(code graph.ErrorCode) int {
	switch code {
	case graph.ErrCodeBadRequest:
		return 400
	case graph.ErrCodeUnauthorized:
		return 401
	case graph.ErrCodeForbidden:
		return 403
	case graph.ErrCodeNotFound:
		return 404
	default:
		return 500
	}
}
