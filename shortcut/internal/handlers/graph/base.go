package graphhandler

import (
	rungraph "github.com/Gadzet005/shortcut/shortcut/internal/usecases/run-graph"
	"go.uber.org/zap"
)

func NewHandlerBase(
	logger *zap.Logger,
	runGraphUC rungraph.UseCase,
) handlerBase {
	return handlerBase{
		logger:     logger,
		runGraphUC: runGraphUC,
	}
}

type handlerBase struct {
	logger     *zap.Logger
	runGraphUC rungraph.UseCase
}
