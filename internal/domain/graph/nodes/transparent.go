package graphnodes

import (
	"context"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	"go.uber.org/zap"
)

var _ graph.NodeExecutor = transparentNodeExecutor{}

func NewTransparentNodeExecutor() graph.NodeExecutor {
	return transparentNodeExecutor{}
}

type transparentNodeExecutor struct{}

func (e transparentNodeExecutor) Run(
	_ context.Context,
	_ *zap.Logger,
	req graph.NodeExecutorRequest,
) (graph.NodeExecutorResponse, error) {
	return graph.NodeExecutorResponse{Items: req.Items}, nil
}

func (e transparentNodeExecutor) TryRevert(
	_ context.Context,
	_ *zap.Logger,
	requestID string,
) (bool, error) {
	return true, nil
}
