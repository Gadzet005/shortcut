package graphnodes

import (
	"context"
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	"go.uber.org/zap"
)

var _ graph.NodeExecutor = metricsExecutor{}

func NewMetricsExecutor(
	inner graph.NodeExecutor,
	nodeID graph.NodeID,
	nodeType string,
	metrics NodeMetrics,
) graph.NodeExecutor {
	return metricsExecutor{
		inner:    inner,
		nodeID:   nodeID,
		nodeType: nodeType,
		metrics:  metrics,
	}
}

type metricsExecutor struct {
	inner    graph.NodeExecutor
	nodeID   graph.NodeID
	nodeType string
	metrics  NodeMetrics
}

func (e metricsExecutor) Run(
	ctx context.Context,
	logger *zap.Logger,
	req graph.NodeExecutorRequest,
) (graph.NodeExecutorResponse, error) {
	start := time.Now()
	resp, err := e.inner.Run(ctx, logger, req)
	duration := time.Since(start)

	if e.metrics != nil {
		e.metrics.ObserveRun(e.nodeID, e.nodeType, duration, err)
	}

	return resp, err
}

func (e metricsExecutor) TryRevert(
	ctx context.Context,
	logger *zap.Logger,
	requestID string,
) (bool, error) {
	return e.inner.TryRevert(ctx, logger, requestID)
}
