package strategy

import (
	"context"

	"github.com/Gadzet005/shortcut/internal/domain/failure"
	"go.uber.org/zap"
)

type saveHandler struct {
	repo   failure.Repo
	logger *zap.Logger
}

func (h saveHandler) Handle(ctx context.Context, f failure.Failure) error {
	f.Status = failure.StatusFailed
	if err := h.repo.Save(ctx, f); err != nil {
		h.logger.Error("save failure record failed", zap.String("request_id", f.RequestID), zap.Error(err))
		return err
	}
	return nil
}
