package graphnodes

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

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

	interval := e.endpoint.InitialInterval
	var lastErr error

	for attempt := 0; attempt <= e.endpoint.RetriesNum; attempt++ {
		if attempt > 0 {
			logger.Warn("retrying node request",
				zap.Int("attempt", attempt),
				zap.Int("retries_num", e.endpoint.RetriesNum),
				zap.Duration("backoff", interval),
				zap.Error(lastErr),
			)
			select {
			case <-ctx.Done():
				return graph.NodeExecutorResponse{}, errors.Wrap(lastErr, "context cancelled during retry backoff")
			case <-time.After(interval):
			}
			interval = nextInterval(interval, e.endpoint.BackoffMultiplier, e.endpoint.MaxInterval)
		}

		resp, err := e.doRequest(ctx, formData)
		if err == nil {
			resp.Meta = map[string]any{
				"status_code": http.StatusOK,
				"retry_count": attempt,
			}
			return resp, nil
		}

		var nodeErr *graph.NodeError
		if errors.As(err, &nodeErr) && nodeErr.Code != graph.ErrCodeInternal {
			return graph.NodeExecutorResponse{}, err
		}

		if ctx.Err() != nil {
			return graph.NodeExecutorResponse{}, err
		}

		lastErr = err
	}

	return graph.NodeExecutorResponse{}, lastErr
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

	return graph.NodeExecutorResponse{Items: results}, nil
}

func nextInterval(current time.Duration, multiplier float64, max time.Duration) time.Duration {
	next := time.Duration(float64(current) * multiplier)
	if max > 0 && next > max {
		return max
	}
	return next
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
