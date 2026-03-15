package di

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Gadzet005/shortcut/pkg/app/config"
	"github.com/Gadzet005/shortcut/pkg/app/env"
	"github.com/Gadzet005/shortcut/pkg/app/lifecycle"
	"github.com/Gadzet005/shortcut/pkg/errors"
	"github.com/Gadzet005/shortcut/pkg/optional"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type service interface {
	Name() string
	Run(ctx lifecycle.Context) error
}

func NewContainer[Config any](service service) *Container[Config] {
	return &Container[Config]{
		service:     service,
		httpServers: make(map[string]*http.Server),
	}
}

type Container[Config any] struct {
	service service
	ctx     lifecycle.Context

	env         optional.T[env.Env]
	config      optional.T[Config]
	logger      optional.T[*zap.Logger]
	httpServers map[string]*http.Server
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

	if err = c.addMetricsServer(); err != nil {
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
	var writers []zapcore.WriteSyncer
	writers = append(writers, zapcore.AddSync(os.Stdout))

	provider, ok := any(c.Config()).(LogConfigProvider)
	if !ok {
		panic(errors.Error("config must embed log config"))
	}
	cfg := provider.GetLogConfig()

	if cfg.Path != "" {
		logDir := filepath.Dir(cfg.Path)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			panic(errors.WrapFail(err, "failed to create log directory"))
		}

		fileWriter := &lumberjack.Logger{
			Filename:   cfg.Path,
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
			LocalTime:  true,
		}
		writers = append(writers, zapcore.AddSync(fileWriter))
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.NewMultiWriteSyncer(writers...),
		zap.InfoLevel,
	)

	if c.Env().IsDev() {
		developmentConfig := zap.NewDevelopmentConfig()
		developmentConfig.OutputPaths = []string{"stdout"}
		if cfg.Path != "" {
			developmentConfig.OutputPaths = append(developmentConfig.OutputPaths, cfg.Path)
		}
		logger, err := developmentConfig.Build()
		if err != nil {
			panic(errors.WrapFail(err, "failed to build development config"))
		}
		return logger
	}

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	return logger
}

// автоматическе запускает http сервер
func (c *Container[Config]) HTTP(name string) *gin.Engine {
	cfg, ok := any(c.Config()).(HTTPConfigProvider)
	if !ok {
		panic(errors.Error("config must embed http config"))
	}

	engine := gin.New()
	c.addHTTPServer(name, cfg.GetHTTPConfig(), engine)

	return engine
}

func (c *Container[Config]) addHTTPServer(name string, cfg HTTPConfig, handler http.Handler) {
	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: handler,
	}
	c.httpServers[name] = &srv
}

func (c *Container[Config]) runHTTPServers(ctx lifecycle.Context) error {
	for name, srv := range c.httpServers {
		logger := c.Logger().Named("http_" + name)
		ctx.RunJob(func(ctx context.Context) {
			logger.Info("http server started", zap.String("addr", srv.Addr))
			err := srv.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Error("http server error", zap.Error(err))
			} else {
				logger.Info("http server stopped")
			}
		}, srv.Shutdown)
	}
	return nil
}

func (c *Container[Config]) addMetricsServer() error {
	cfg, ok := any(c.Config()).(MetricsConfigProvider)
	if !ok {
		return nil
	}

	name := c.service.Name() + "_metrics"
	c.addHTTPServer(name, cfg.GetMetricsConfig().HTTPConfig, promhttp.Handler())
	return nil
}
