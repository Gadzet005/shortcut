package lifecycle

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	errorsutils "github.com/Gadzet005/shortcut/shortcut/pkg/utils/errors"
)

const stopTimeOut = 10 * time.Second

func Run(c Cycler) error {
	ctx := context.Background()
	if err := c.Start(ctx); err != nil {
		return errorsutils.WrapFail(err, "start cycler")
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctxWithTimeOut, cancel := context.WithTimeout(ctx, stopTimeOut)
	defer cancel()

	if err := c.Stop(ctxWithTimeOut); err != nil {
		return errorsutils.WrapFail(err, "stop cycler")
	}
	return nil
}
