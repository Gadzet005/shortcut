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
	if e.endpoint.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, e.endpoint.Timeout)
		defer cancel()
	}

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

	return graph.NodeExecutorResponse{Items: results}, nil
}

func httpStatusToErrorCode(status int) graph.ErrorCode {
	switch status {
	case 400:
		return graph.ErrCodeBadRequest
	case 401:
		return graph.ErrCodeUnauthorized
	case 403:
		return graph.ErrCodeForbidden
	case 404:
		return graph.ErrCodeNotFound
	default:
		return graph.ErrCodeInternal
	}
}
