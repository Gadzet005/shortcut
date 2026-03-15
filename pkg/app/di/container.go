package di

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Gadzet005/shortcut/pkg/app/config"
	"github.com/Gadzet005/shortcut/pkg/app/env"
	"github.com/Gadzet005/shortcut/pkg/app/lifecycle"
	"github.com/Gadzet005/shortcut/pkg/errors"
	"github.com/Gadzet005/shortcut/pkg/optional"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type service interface {
	Name() string
	Run(ctx lifecycle.Context) error
}

func NewContainer[Config any](service service) *Container[Config] {
	return &Container[Config]{
		service:     service,
		httpServers: make(map[string]httpServer),
	}
}

type Container[Config any] struct {
	service service
	ctx     lifecycle.Context

	env         optional.T[env.Env]
	config      optional.T[Config]
	logger      optional.T[*zap.Logger]
	httpServers map[string]httpServer
}

type httpServer struct {
	Server *http.Server
	Engine *gin.Engine
}

func (c *Container[Config]) run(ctx lifecycle.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				panic(r)
			}
		}
	}()

	c.ctx = ctx
	if err = c.service.Run(ctx); err != nil {
		return err
	}

	return c.runHTTPServers(ctx)
}

func (c *Container[Config]) Env() env.Env {
	if v, ok := c.env.Value(); ok {
		return v
	}

	env := env.LoadFromEnvVar()
	c.env = optional.New(env)
	return env
}

func (c *Container[Config]) Config() Config {
	if v, ok := c.config.Value(); ok {
		return v
	}

	_ = godotenv.Load()

	config, err := config.LoadServiceConfig[Config](c.service.Name(), c.Env())
	if err != nil {
		panic(errors.WrapFail(err, "load config"))
	}

	c.config = optional.New(config)
	return config
}

func (c *Container[Config]) Logger() *zap.Logger {
	if v, ok := c.logger.Value(); ok {
		return v
	}

	var logger *zap.Logger
	var err error
	if c.Env() == env.EnvProd {
		logger, err = zap.NewProduction(zap.AddStacktrace(zap.FatalLevel))
	} else {
		logger, err = zap.NewDevelopment(zap.AddStacktrace(zap.FatalLevel))
	}
	if err != nil {
		panic(errors.WrapFail(err, "create logger"))
	}

	c.logger = optional.New(logger)
	return logger
}

// автоматическе запускает http сервер
func (c *Container[Config]) HTTP(name string) *gin.Engine {
	if _, ok := c.httpServers[name]; ok {
		return c.httpServers[name].Engine
	}

	cfg, ok := any(c.Config()).(HTTPConfigProvider)
	if !ok {
		panic(errors.Error("config must embed http config"))
	}

	engine := gin.New()

	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.GetHTTPConfig().Port),
		Handler: engine,
	}
	c.httpServers[name] = httpServer{
		Server: &srv,
		Engine: engine,
	}

	return engine
}

func (c *Container[Config]) runHTTPServers(ctx lifecycle.Context) error {
	for name, srv := range c.httpServers {
		logger := c.Logger().Named("http_" + name)
		ctx.RunJob(func(ctx context.Context) {
			logger.Info("http server started")
			err := srv.Server.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Error("http server error", zap.Error(err))
			} else {
				logger.Info("http server stopped")
			}
		}, srv.Server.Shutdown)
	}
	return nil
}
