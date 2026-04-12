package graphhandler

import (
	rungraph "github.com/Gadzet005/shortcut/internal/usecases/run-graph"
)

func NewHandlerBase(
	runGraphUC rungraph.UseCase,
) handlerBase {
	return handlerBase{
		runGraphUC: runGraphUC,
	}
}

type handlerBase struct {
	runGraphUC rungraph.UseCase
}
