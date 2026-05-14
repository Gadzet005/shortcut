package strategy

import (
	"context"

	"github.com/Gadzet005/shortcut/internal/domain/failure"
	"go.uber.org/zap"
)

type absentHandler struct {
	logger *zap.Logger
}

func (h absentHandler) Handle(_ context.Context, f failure.Failure) error {
	h.logger.Debug("absent strategy applied", zap.String("request_id", f.RequestID))
	return nil
}
