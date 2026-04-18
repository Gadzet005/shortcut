package revertrequest

import (
	rungraph "github.com/Gadzet005/shortcut/internal/usecases/run-graph"
	revertrequest "github.com/Gadzet005/shortcut/internal/usecases/revert-request"
)

func NewHandlerBase(
	runGraphUC rungraph.UseCase,
	revertRequestUC revertrequest.UseCase,
	tracingEnabled bool,
) handlerBase {
	return handlerBase{
		runGraphUC:      runGraphUC,
		revertRequestUC: revertRequestUC,
		tracingEnabled:  tracingEnabled,
	}
}

type handlerBase struct {
	runGraphUC      rungraph.UseCase
	revertRequestUC revertrequest.UseCase
	tracingEnabled  bool
}