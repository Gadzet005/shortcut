package lifecycle

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/Gadzet005/shortcut/pkg/errors"
)

func Run(app App) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	appCtx := NewContext(ctx)
	if err := app.OnRun(appCtx); err != nil {
		app.OnRunFailed(appCtx, err)
		return
	}

	<-ctx.Done()

	app.OnStopStarted(appCtx)

	shutdownCtx, shutdownCancel := app.ShutdownContext(appCtx)
	defer shutdownCancel()

	closedCh := make(chan struct{})
	go func() {
		appCtx.wg.Wait()
		close(closedCh)
	}()

	var errs []error
	for _, stopper := range appCtx.stoppers {
		if err := stopper(shutdownCtx); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		app.OnStopFailed(appCtx, errors.WrapFail(errors.Join(errs...), "stop jobs"))
		return
	}

	select {
	case <-closedCh:
		app.OnStopCompleted(appCtx)
		return
	case <-shutdownCtx.Done():
		app.OnStopFailed(
			appCtx,
			errors.WrapFail(
				fmt.Errorf("shutdown timeout: %v", shutdownCtx.Err()),
				"shutdown timeout",
			),
		)
		return
	}
}
