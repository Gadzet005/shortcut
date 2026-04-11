package graphhandler

import (
	rungraph "github.com/Gadzet005/shortcut/internal/usecases/run-graph"
)

func NewHandlerBase(
	runGraphUC rungraph.UseCase,
	tracingEnabled bool,
) handlerBase {
	return handlerBase{
		runGraphUC:     runGraphUC,
		tracingEnabled: tracingEnabled,
	}
}

type handlerBase struct {
	runGraphUC     rungraph.UseCase
	tracingEnabled bool
}
