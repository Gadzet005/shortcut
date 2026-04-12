package graphnodes

import (
	"context"
	"encoding/json"
	"net/http"

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

	endpoint := applyEndpointOverride(e.endpoint, req.EndpointOverride)
	return withRetry(ctx, logger, endpoint, func(ctx context.Context) (graph.NodeExecutorResponse, error) {
		return e.doRequest(ctx, endpoint, formData)
	})
}

func (e defaultNodeExecutor) doRequest(
	ctx context.Context,
	endpoint Endpoint,
	formData map[string]string,
) (graph.NodeExecutorResponse, error) {
	reqCtx := ctx
	if endpoint.Timeout > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(ctx, endpoint.Timeout)
		defer cancel()
	}

	resp, err := e.client.R().
		SetContext(reqCtx).
		SetFormData(formData).
		SetDoNotParseResponse(true).
		Post(endpoint.URL)
	if err != nil {
		return graph.NodeExecutorResponse{}, errors.Wrap(err, "make request")
	}
	if resp.StatusCode() != http.StatusOK {
		errorResponse := make(map[string]any)
		if err := json.NewDecoder(resp.RawResponse.Body).Decode(&errorResponse); err != nil {
			return graph.NodeExecutorResponse{}, errors.Wrap(err, "decode error response")
		}
		return graph.NodeExecutorResponse{}, &graph.NodeError{
			Code:    httpStatusToErrorCode(resp.StatusCode()),
			Payload: errorResponse,
		}
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

	return graph.NodeExecutorResponse{
		Items: results,
		Meta:  map[string]any{"status_code": http.StatusOK},
	}, nil
}

