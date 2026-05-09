package strategy

import (
	"context"

	"github.com/Gadzet005/shortcut/internal/domain/failure"
)

type absentHandler struct{}

func (absentHandler) Handle(_ context.Context, _ failure.Failure) error {
	return nil
}
