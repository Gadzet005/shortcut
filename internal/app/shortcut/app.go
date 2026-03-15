package app

import (
	"flag"
	"log"

	graphconfig "github.com/Gadzet005/shortcut/internal/domain/graph/config"
	graphhandler "github.com/Gadzet005/shortcut/internal/handlers/graph"
	"github.com/Gadzet005/shortcut/internal/middleware"
	graphrepolocal "github.com/Gadzet005/shortcut/internal/repo/graph/local"
	rungraph "github.com/Gadzet005/shortcut/internal/usecases/run-graph"
	"github.com/Gadzet005/shortcut/pkg/app/config"
	"github.com/Gadzet005/shortcut/pkg/app/di"
	"github.com/Gadzet005/shortcut/pkg/app/lifecycle"
	"github.com/Gadzet005/shortcut/pkg/errors"
	"github.com/Gadzet005/shortcut/pkg/metrics"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

func NewService() lifecycle.App {
	s := &service{}
	s.Container = di.NewContainer[Config](s)
	return di.NewApp(s.Container)
}

type service struct {
	*di.Container[Config]
}

func (s service) Name() string {
	return "shortcut"
}

func (s service) Run(c lifecycle.Context) error {
	graphConfigPath := flag.String(
		"graphconfigs",
		"./tests/e2e/configs/graph.yaml",
		"path to graph config file",
	)
	flag.Parse()

	if s.Env().IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}

	serviceMetrics := metrics.NewHTTPServiceMetrics("shortcut")

	graphConfig, err := config.Load[graphconfig.Config](*graphConfigPath)
	if err != nil {
		log.Fatalf("failed to load graph config: %v", err)
	}

	graphConfigs, err := graphconfig.ParseConfig(graphConfig.Namespace, func(msg string) {
		s.Logger().Warn(msg)
	})
	if err != nil {
		return errors.WrapFailf(err, "failed to parse graph config")
	}
	localRepo := graphrepolocal.NewLocalRepo(graphConfigs)

	r := s.HTTP("shortcut")
	r.Use(
		middleware.ZapLogger(s.Logger()),
		middleware.ZapRecovery(s.Logger(), true),
		serviceMetrics.MetricsMiddleware(),
	)

	client := resty.New()
	runGraphUC := rungraph.NewUseCase(client, s.Logger(), localRepo)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	{
		handlerBase := graphhandler.NewHandlerBase(s.Logger(), runGraphUC)

		g := r.Group("/graph")
		g.POST("/:graph_id/run", handlerBase.RunGraph)
	}

	return nil
}
