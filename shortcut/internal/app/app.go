package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	graphhandler "github.com/Gadzet005/shortcut/shortcut/internal/handlers/graph"
	"github.com/Gadzet005/shortcut/shortcut/internal/middleware"
	graphrepostub "github.com/Gadzet005/shortcut/shortcut/internal/repo/graph/stub"
	rungraph "github.com/Gadzet005/shortcut/shortcut/internal/usecases/run-graph"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

func NewService(config Config, serviceConfig graphconfig.Config,  logger *zap.Logger) (service, error) {
	if config.Env.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}

	graphConfigs, err := graphconfig.ParseConfig(serviceConfig.Namespace, func (s string) {
		logger.Info(s)
	})
	if err != nil {
		return service{}, errorsutils.WrapFail(err, "failed to setup services: %s", err.Error)
	}
	
	logger.Info("Start with config of graphs", zap.String("config", fmt.Sprintf("%v", graphConfigs)))

	repo := graphlocalrepo.NewLocalRepo(graphConfigs)
	if err != nil {
		return service{}, errorsutils.WrapFail(err, "failed to create repo: %s", err.Error)
	}

	r := gin.New()
	r.Use(middleware.ZapLogger(logger))
	r.Use(middleware.ZapRecovery(logger, true))

	client := resty.New()
	graphRepo := graphrepostub.NewStubRepo()
	runGraphUC := rungraph.NewUseCase(client, logger, graphRepo)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	{
		handlerBase := graphhandler.NewHandlerBase(logger, runGraphUC)

		g := r.Group("/graph")
		g.POST("/:graph_id/run", handlerBase.RunGraph)
	}

	return service{
		logger: logger,
		srv: &http.Server{
			Addr:              fmt.Sprintf(":%d", config.HTTPServer.Port),
			Handler:           r,
			ReadHeaderTimeout: 10 * time.Second,
		},
	}, nil
}

type service struct {
	logger *zap.Logger
	srv    *http.Server
}

func (s service) Start(ctx context.Context) error {
	go func() {
		s.logger.Info("starting server", zap.String("addr", s.srv.Addr))
		err := s.srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("server stopped", zap.Error(err))
		}
		s.logger.Info("server stopped")
	}()

	return nil
}

func (s service) Stop(ctx context.Context) error {
	s.logger.Info("stopping server")
	return s.srv.Shutdown(ctx)
}
