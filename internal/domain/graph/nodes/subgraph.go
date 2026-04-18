package graphnodes

import (
	"context"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	errors "github.com/Gadzet005/shortcut/pkg/errors"
	"go.uber.org/zap"
)

var _ graph.NodeExecutor = subgraphNodeExecutor{}

func NewSubgraphNodeExecutor(graph graph.Graph) graph.NodeExecutor {
	return subgraphNodeExecutor{
		graph: graph,
	}
}

type subgraphNodeExecutor struct {
	graph graph.Graph
}

func (e subgraphNodeExecutor) Run(
	ctx context.Context,
	logger *zap.Logger,
	req graph.NodeExecutorRequest,
) (graph.NodeExecutorResponse, error) {
	results, err := e.graph.Run(ctx, logger, req.Items)
	if err != nil {
		return graph.NodeExecutorResponse{}, errors.Wrap(err, "run graph")
	}
	return graph.NodeExecutorResponse{Items: results}, nil
}

func (e subgraphNodeExecutor) TryRevert(
	ctx context.Context,
	logger *zap.Logger,
	requestID string,
) (bool, error) {
	revertResult, err := e.graph.TryRevert(ctx, logger, requestID)
	if err != nil {
		return false, errors.Wrap(err, "revert graph")
	}
	return revertResult, nil
}
