package graphnodes

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	errors "github.com/Gadzet005/shortcut/pkg/errors"
	shortcutapi "github.com/Gadzet005/shortcut/pkg/shortcut/api"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const (
	httpAdapterRequestItemID  graph.ItemID = "http_request"
	httpAdapterResponseItemID graph.ItemID = "http_response"
)

var _ graph.NodeExecutor = httpAdapterNodeExecutor{}

func NewHTTPAdapterNodeExecutor(client *resty.Client, endpoint Endpoint) graph.NodeExecutor {
	return httpAdapterNodeExecutor{
		endpoint: endpoint,
		client:   client,
	}
}

type httpAdapterNodeExecutor struct {
	endpoint Endpoint
	client   *resty.Client
}

func (e httpAdapterNodeExecutor) Run(
	ctx context.Context,
	logger *zap.Logger,
	req graph.NodeExecutorRequest,
) (graph.NodeExecutorResponse, error) {
	httpReqItem, ok := req.Items[httpAdapterRequestItemID]
	if !ok {
		return graph.NodeExecutorResponse{}, errors.Error("http_request item not found")
	}

	var httpReq shortcutapi.HttpRequest
	if err := json.Unmarshal(httpReqItem.Data, &httpReq); err != nil {
		return graph.NodeExecutorResponse{}, errors.Wrap(err, "unmarshal http_request")
	}

	return withRetry(ctx, logger, e.endpoint, func(ctx context.Context) (graph.NodeExecutorResponse, error) {
		return e.doRequest(ctx, httpReq)
	})
}

func (e httpAdapterNodeExecutor) doRequest(
	ctx context.Context,
	httpReq shortcutapi.HttpRequest,
) (graph.NodeExecutorResponse, error) {
	reqCtx := ctx
	if e.endpoint.Timeout > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(ctx, e.endpoint.Timeout)
		defer cancel()
	}

	r := e.client.R().
		SetContext(reqCtx).
		SetHeaderMultiValues(httpReq.Headers).
		SetQueryParamsFromValues(httpReq.Query).
		SetDoNotParseResponse(true)

	if len(httpReq.Body) > 0 {
		r = r.SetBody(httpReq.Body)
	}

	resp, err := r.Execute(httpReq.Method, e.endpoint.URL)
	if err != nil {
		return graph.NodeExecutorResponse{}, errors.Wrap(err, "make request")
	}
	defer resp.RawResponse.Body.Close()

	body, err := io.ReadAll(resp.RawResponse.Body)
	if err != nil {
		return graph.NodeExecutorResponse{}, errors.Wrap(err, "read response body")
	}

	statusCode := resp.StatusCode()
	if statusCode >= http.StatusInternalServerError {
		payload := make(map[string]any)
		_ = json.Unmarshal(body, &payload)
		return graph.NodeExecutorResponse{}, &graph.NodeError{
			Code:    graph.ErrCodeInternal,
			Payload: payload,
		}
	}
	if statusCode >= http.StatusBadRequest {
		payload := make(map[string]any)
		_ = json.Unmarshal(body, &payload)
		return graph.NodeExecutorResponse{}, &graph.NodeError{
			Code:    httpStatusToErrorCode(statusCode),
			Payload: payload,
		}
	}

	httpResp := shortcutapi.HttpResponse{
		StatusCode: statusCode,
		Headers:    resp.Header(),
		Body:       body,
	}

	respData, err := json.Marshal(httpResp)
	if err != nil {
		return graph.NodeExecutorResponse{}, errors.Wrap(err, "marshal http_response")
	}

	return graph.NodeExecutorResponse{
		Items: map[graph.ItemID]graph.Item{
			httpAdapterResponseItemID: {Data: respData},
		},
		Meta: map[string]any{"status_code": statusCode},
	}, nil
}
