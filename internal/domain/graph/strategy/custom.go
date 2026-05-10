package strategy

import (
	"context"
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/failure"
	"github.com/Gadzet005/shortcut/internal/domain/graph"
	"go.uber.org/zap"
)

type customHandler struct {
	repo   failure.Repo
	steps  []graph.FailureStep
	logger *zap.Logger
}

func (h customHandler) Handle(ctx context.Context, f failure.Failure) error {
	h.logger.Info("custom strategy started",
		zap.String("request_id", f.RequestID),
		zap.String("namespace_id", f.NamespaceID),
		zap.String("graph_id", f.GraphID),
		zap.Int("steps", len(h.steps)),
	)

	f.NumRetry = 0
	f.Status = failure.StatusPending

	var firstDelay time.Duration
	if len(h.steps) > 0 {
		firstDelay = h.steps[0].WaitBefore
	}
	f.ReadyToRetryAt = time.Now().Add(firstDelay)

	if err := h.repo.Save(ctx, f); err != nil {
		h.logger.Error("save custom failure record failed", zap.String("request_id", f.RequestID), zap.Error(err))
		return err
	}

	h.logger.Info("custom strategy finished",
		zap.String("request_id", f.RequestID),
		zap.Duration("first_delay", firstDelay),
	)
	return nil
}
