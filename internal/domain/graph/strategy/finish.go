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
	h.logger.Info("finish started",
		zap.String("request_id", f.RequestID),
		zap.String("namespace_id", f.NamespaceID),
		zap.String("graph_id", f.GraphID),
	)

	ok, err := h.recovery.Finish(ctx, f.NamespaceID, f.GraphID, f.RequestID, f.VisitedNodes(), f.Method, f.Path, f.RequestBody)
	if err != nil {
		h.logger.Error("finish failed", zap.String("request_id", f.RequestID), zap.Error(err))
		return err
	}
	if !ok {
		h.logger.Warn("finish did not complete graph", zap.String("request_id", f.RequestID))
		return nil
	}

	h.logger.Info("finish finished", zap.String("request_id", f.RequestID))
	return nil
}
