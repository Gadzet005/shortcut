package strategy

import (
	"context"

	"github.com/Gadzet005/shortcut/internal/domain/failure"
	"go.uber.org/zap"
)

type revertHandler struct {
	recovery failure.Recovery
	logger   *zap.Logger
}

func (h revertHandler) Handle(ctx context.Context, f failure.Failure) error {
	ok, err := h.recovery.Revert(ctx, f.NamespaceID, f.GraphID, f.RequestID, f.VisitedNodes())
	if err != nil {
		h.logger.Error("revert failed", zap.String("request_id", f.RequestID), zap.Error(err))
		return err
	}
	if !ok {
		h.logger.Warn("revert partially failed", zap.String("request_id", f.RequestID))
	}
	return nil
}
