package graphnodes

import (
	"context"
	"net/http"
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	errors "github.com/Gadzet005/shortcut/pkg/errors"
	"go.uber.org/zap"
)

func withRetry(
	ctx context.Context,
	logger *zap.Logger,
	endpoint Endpoint,
	fn func(ctx context.Context) (graph.NodeExecutorResponse, error),
) (graph.NodeExecutorResponse, error) {
	interval := endpoint.InitialInterval
	var lastErr error

	for attempt := 0; attempt <= endpoint.RetriesNum; attempt++ {
		if attempt > 0 {
			logger.Warn("retrying node request",
				zap.Int("attempt", attempt),
				zap.Int("retries_num", endpoint.RetriesNum),
				zap.Duration("backoff", interval),
				zap.Error(lastErr),
			)
			select {
			case <-ctx.Done():
				return graph.NodeExecutorResponse{}, errors.Wrap(lastErr, "context cancelled during retry backoff")
			case <-time.After(interval):
			}
			interval = nextInterval(interval, endpoint.BackoffMultiplier, endpoint.MaxInterval)
		}

		resp, err := fn(ctx)
		if err == nil {
			if resp.Meta == nil {
				resp.Meta = make(map[string]any)
			}
			resp.Meta["retry_count"] = attempt
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

func nextInterval(current time.Duration, multiplier float64, max time.Duration) time.Duration {
	next := time.Duration(float64(current) * multiplier)
	if max > 0 && next > max {
		return max
	}
	return next
}

func httpStatusToErrorCode(status int) graph.ErrorCode {
	switch status {
	case http.StatusBadRequest:
		return graph.ErrCodeBadRequest
	case http.StatusUnauthorized:
		return graph.ErrCodeUnauthorized
	case http.StatusForbidden:
		return graph.ErrCodeForbidden
	case http.StatusNotFound:
		return graph.ErrCodeNotFound
	default:
		return graph.ErrCodeInternal
	}
}
