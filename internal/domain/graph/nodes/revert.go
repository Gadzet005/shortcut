package graphnodes

import (
	"context"
	"net/http"

	errors "github.com/Gadzet005/shortcut/pkg/errors"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const revertRequestIDKey = "request_id"

func tryRevertHTTP(
	ctx context.Context,
	logger *zap.Logger,
	client *resty.Client,
	endpoint Endpoint,
	requestID string,
) (bool, error) {
	if endpoint.RevertURL == "" {
		return true, nil
	}

	reqCtx := ctx
	if endpoint.Timeout > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(ctx, endpoint.Timeout)
		defer cancel()
	}

	resp, err := client.R().
		SetContext(reqCtx).
		SetFormData(map[string]string{revertRequestIDKey: requestID}).
		Post(endpoint.RevertURL)
	if err != nil {
		logger.Warn("revert request failed", zap.String("url", endpoint.RevertURL), zap.Error(err))
		return false, errors.Wrap(err, "make revert request")
	}

	status := resp.StatusCode()
	if status >= http.StatusBadRequest && status < http.StatusInternalServerError {
		logger.Warn("revert endpoint returned client error",
			zap.String("url", endpoint.RevertURL),
			zap.Int("status_code", status))
		return false, nil
	}
	if status >= http.StatusInternalServerError {
		return false, errors.Errorf("revert endpoint returned status %d", status)
	}
	return true, nil
}
