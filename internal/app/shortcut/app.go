package app

import (
	"flag"

	graphconfig "github.com/Gadzet005/shortcut/internal/domain/graph/config"
	graphhandler "github.com/Gadzet005/shortcut/internal/handlers/graph"
	graphlocalrepo "github.com/Gadzet005/shortcut/internal/repo/graph/local"
	rungraph "github.com/Gadzet005/shortcut/internal/usecases/run-graph"
	"github.com/Gadzet005/shortcut/pkg/app/di"
	"github.com/Gadzet005/shortcut/pkg/app/lifecycle"
	"github.com/Gadzet005/shortcut/pkg/errors"
	httpmiddleware "github.com/Gadzet005/shortcut/pkg/http/middleware"
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
		"./tests/configs",
		"path to graph config file",
	)
	flag.Parse()

	if s.Env().IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}

	cfg, err := graphconfig.Load(*graphConfigPath)
	if err != nil {
		return errors.WrapFailf(err, "failed to load graph config")
	}
	client := resty.New()
	graphConfig, err := graphconfig.Convert(cfg, func(msg string) {
		s.Logger().Warn(msg)
	}, client)
	if err != nil {
		return errors.WrapFailf(err, "failed to convert graph config")
	}
	localRepo := graphlocalrepo.NewLocalRepo(graphConfig)

	r := s.HTTP("shortcut")
	r.Use(
		httpmiddleware.ZapLogger(s.Logger()),
		httpmiddleware.Metrics("shortcut"),
		httpmiddleware.Recover(),
	)

	runGraphUC := rungraph.NewUseCase(client, s.Logger(), localRepo)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	handlerBase := graphhandler.NewHandlerBase(runGraphUC)
	r.Any("run/:namespace_id/*path", handlerBase.RunGraph)

	return nil
}
