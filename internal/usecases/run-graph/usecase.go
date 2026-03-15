package rungraph

import (
	"context"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	"github.com/Gadzet005/shortcut/pkg/errors"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

func NewUseCase(
	client *resty.Client,
	logger *zap.Logger,
	graphRepo graph.Repo,
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
	graphRepo graph.Repo
}

func (u useCase) RunGraph(ctx context.Context, input Request) (Response, error) {
	g, err := u.graphRepo.GetGraph(input.GraphID)
	if err != nil {
		return Response{}, errors.Wrap(err, "get graph")
	}

	resp, err := g.Run(ctx, u.logger, graph.RunNodeRequest{
		Client: u.client,
		Items: map[graph.ItemID]graph.Item{
			graph.DefaultItemID: {Data: input.Data},
		},
	})
	if err != nil {
		return Response{}, errors.Wrap(err, "run graph")
	}

	item, ok := resp.Items[graph.DefaultItemID]
	if !ok {
		return Response{}, errors.Wrap(err, "get item")
	}
	return Response{
		Data: item.Data,
	}, nil
}
