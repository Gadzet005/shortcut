package graphnodes

import (
	"context"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	errors "github.com/Gadzet005/shortcut/pkg/errors"
	multipartutils "github.com/Gadzet005/shortcut/pkg/utils/multipart"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

var _ graph.NodeExecutor = defaultNodeExecutor{}

func NewDefaultNodeExecutor(client *resty.Client, endpoint Endpoint) graph.NodeExecutor {
	return defaultNodeExecutor{
		endpoint: endpoint,
		client:   client,
	}
}

type defaultNodeExecutor struct {
	endpoint Endpoint
	client   *resty.Client
}

func (e defaultNodeExecutor) Run(
	ctx context.Context,
	logger *zap.Logger,
	req graph.NodeExecutorRequest,
) (graph.NodeExecutorResponse, error) {
	formData := make(map[string]string, len(req.Items))
	for id, item := range req.Items {
		formData[id.String()] = string(item.Data)
	}

	resp, err := e.client.R().
		SetContext(ctx).
		SetFormData(formData).
		SetDoNotParseResponse(true).
		Post(e.endpoint.URL)
	if err != nil {
		return graph.NodeExecutorResponse{}, errors.Wrap(err, "make request")
	}
	if !resp.IsSuccess() {
		return graph.NodeExecutorResponse{}, errors.Errorf(
			"request failed with status code %d and body %s",
			resp.StatusCode(), resp.String(),
		)
	}

	body := resp.RawResponse.Body
	defer body.Close()

	data, err := multipartutils.ReadMultipartData(resp.Header(), body)
	if err != nil {
		return graph.NodeExecutorResponse{}, errors.Wrap(err, "read multipart data")
	}

	results := make(map[graph.ItemID]graph.Item, len(data))
	for id, item := range data {
		results[graph.ItemID(id)] = graph.Item{Data: item}
	}

	return graph.NodeExecutorResponse{Items: results}, nil
}
