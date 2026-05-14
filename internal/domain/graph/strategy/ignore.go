package strategy

import (
	"context"

	"github.com/Gadzet005/shortcut/internal/domain/failure"
	"go.uber.org/zap"
)

type ignoreHandler struct {
	logger *zap.Logger
}

func (h ignoreHandler) Handle(_ context.Context, f failure.Failure) error {
	h.logger.Debug("ignore strategy applied", zap.String("request_id", f.RequestID))
	return nil
}
