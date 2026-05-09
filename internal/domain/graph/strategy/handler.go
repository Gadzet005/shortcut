package strategy

import (
	"context"

	"github.com/Gadzet005/shortcut/internal/domain/failure"
	"github.com/Gadzet005/shortcut/internal/domain/graph"
	"go.uber.org/zap"
)

type Handler interface {
	Handle(ctx context.Context, f failure.Failure) error
}

type Factory interface {
	New(strategy graph.FailureStrategy, steps []graph.FailureStep) Handler
}

type factory struct {
	repo     failure.Repo
	recovery failure.Recovery
	logger   *zap.Logger
}

func NewFactory(repo failure.Repo, recovery failure.Recovery, logger *zap.Logger) Factory {
	return factory{
		repo:     repo,
		recovery: recovery,
		logger:   logger.Named("failure-strategy"),
	}
}

func (f factory) New(strategy graph.FailureStrategy, steps []graph.FailureStep) Handler {
	switch strategy {
	case graph.RevertFailureStrategy:
		return revertHandler{recovery: f.recovery, logger: f.logger}
	case graph.SaveFailureStrategy:
		return saveHandler{repo: f.repo, logger: f.logger}
	case graph.FinishFailureStrategy:
		return finishHandler{recovery: f.recovery, logger: f.logger}
	case graph.CustomFailureStrategy:
		return customHandler{repo: f.repo, steps: steps, logger: f.logger}
	case graph.IgnoreFailureStrategy:
		return ignoreHandler{}
	default:
		return absentHandler{}
	}
}
