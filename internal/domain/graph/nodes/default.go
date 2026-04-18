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

const (
	defaultRevertResultItemID graph.ItemID = "result"
)

var _ graph.NodeExecutor = defaultNodeExecutor{}

func NewDefaultNodeExecutor(client *resty.Client, endpoint, revertEndpoint Endpoint) graph.NodeExecutor {
	return defaultNodeExecutor{
		endpoint: endpoint,
		revertEndpoint: revertEndpoint,
		client:   client,
	}
}

type defaultNodeExecutor struct {
	endpoint Endpoint
	revertEndpoint Endpoint
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

	return withRetry(ctx, logger, e.endpoint, func(ctx context.Context) (graph.NodeExecutorResponse, error) {
		return e.doRequest(ctx, formData)
	})
}

func (e defaultNodeExecutor) TryRevert(
	ctx context.Context,
	logger *zap.Logger,
	requestID string,
) (bool, error) {
	formData := map[string]string{
		"request_id": requestID,
	}

	response, err := withRetry(ctx, logger, e.revertEndpoint, func(ctx context.Context) (graph.NodeExecutorResponse, error) {
		return e.doRequest(ctx, formData)
	})
	if err != nil {
		return false,  errors.Wrap(err, "do http with retry")
	}

	result, ok := response.Items[defaultRevertResultItemID]
	if !ok {
		return false, errors.Error("result item not found")
	}

	var booleanResult bool
	if err := json.Unmarshal(result.Data, &booleanResult); err != nil {
		return false, errors.Wrap(err, "unmarshal revert result")
	}

	return booleanResult, nil
}

func (e defaultNodeExecutor) doRequest(
	ctx context.Context,
	formData map[string]string,
) (graph.NodeExecutorResponse, error) {
	reqCtx := ctx
	if e.endpoint.Timeout > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(ctx, e.endpoint.Timeout)
		defer cancel()
	}

	resp, err := e.client.R().
		SetContext(reqCtx).
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

	return graph.NodeExecutorResponse{
		Items: results,
		Meta:  map[string]any{"status_code": http.StatusOK},
	}, nil
}

