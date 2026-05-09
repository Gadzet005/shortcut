package strategy

import (
	"context"

	"github.com/Gadzet005/shortcut/internal/domain/failure"
	"go.uber.org/zap"
)

type finishHandler struct {
	recovery failure.Recovery
	logger   *zap.Logger
}

func (h finishHandler) Handle(ctx context.Context, f failure.Failure) error {
	ok, err := h.recovery.Finish(ctx, f.NamespaceID, f.GraphID, f.RequestID, f.VisitedNodes(), f.Method, f.Path, f.RequestBody)
	if err != nil {
		h.logger.Error("finish failed", zap.String("request_id", f.RequestID), zap.Error(err))
		return err
	}
	if !ok {
		h.logger.Warn("finish did not complete graph", zap.String("request_id", f.RequestID))
	}
	return nil
}
