package rungraph

import (
	"context"

	"github.com/Gadzet005/shortcut/shortcut/internal/domain/graph"
	errorsutils "github.com/Gadzet005/shortcut/shortcut/pkg/utils/errors"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

func NewUseCase(
	client *resty.Client,
	logger *zap.Logger,
	graphRepo graph.GraphRepo,
) useCase {
	return useCase{
		client:    client,
		graphRepo: graphRepo,
		logger:    logger,
	}
}

type useCase struct {
	client    *resty.Client
	logger    *zap.Logger
	graphRepo graph.GraphRepo
}

func (u useCase) RunGraph(ctx context.Context, input RunGraphRequest) (RunGraphResponse, error) {
	g, err := u.graphRepo.GetGraph(input.GraphID)
	if err != nil {
		return RunGraphResponse{}, errorsutils.WrapFail(err, "get graph")
	}

	resp, err := g.Run(ctx, u.logger, graph.RunNodeRequest{
		Client: u.client,
		Items: map[graph.ItemID]graph.Item{
			graph.DefaultItemID: {Data: input.Data},
		},
	})
	if err != nil {
		return RunGraphResponse{}, errorsutils.WrapFail(err, "run graph")
	}

	item, ok := resp.Items[graph.DefaultItemName]
	if !ok {
		return RunGraphResponse{}, errorsutils.WrapFail(err, "get item")
	}
	return RunGraphResponse{
		Data: item.Data,
	}, nil
}
