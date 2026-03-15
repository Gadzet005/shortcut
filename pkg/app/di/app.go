package di

import (
	"context"

	"github.com/Gadzet005/shortcut/pkg/app/lifecycle"
	"github.com/Gadzet005/shortcut/pkg/errors"
	"go.uber.org/zap"
)

func NewApp[Config any](container *Container[Config]) *app[Config] {
	cfg, ok := any(container.Config()).(AppConfigProvider)
	if !ok {
		panic(errors.Error("config must embed app config"))
	}
	return &app[Config]{
		container: container,
		cfg:       cfg.GetAppConfig(),
	}
}

type app[Config any] struct {
	container *Container[Config]
	cfg       AppConfig
}

func (a *app[Config]) OnRun(ctx lifecycle.Context) error {
	return a.container.run(ctx)
}

func (a *app[Config]) OnRunFailed(ctx lifecycle.Context, err error) {
	a.container.Logger().Error("run failed", zap.Error(err))
}

func (a *app[Config]) OnStopStarted(ctx lifecycle.Context) {
	a.container.Logger().Info("stopping background jobs...")
}

func (a *app[Config]) OnStopFailed(ctx lifecycle.Context, err error) {
	a.container.Logger().Error("failed to stop background jobs", zap.Error(err))
}

func (a *app[Config]) OnStopCompleted(ctx lifecycle.Context) {
	a.container.Logger().Info("application stopped successfully")
}

func (a *app[Config]) ShutdownContext(ctx lifecycle.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), a.cfg.ShutdownTimeout)
}
