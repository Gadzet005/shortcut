package app

import (
	"context"
	"flag"

	graphconfig "github.com/Gadzet005/shortcut/internal/domain/graph/config"
	"github.com/Gadzet005/shortcut/internal/domain/trace"
	graphhandler "github.com/Gadzet005/shortcut/internal/handlers/graph"
	tracehandler "github.com/Gadzet005/shortcut/internal/handlers/trace"
	graphlocalrepo "github.com/Gadzet005/shortcut/internal/repo/graph/local"
	tracemongo "github.com/Gadzet005/shortcut/internal/repo/trace/mongo"
	failurepostgres "github.com/Gadzet005/shortcut/internal/repo/failure/postgres"
	rungraph "github.com/Gadzet005/shortcut/internal/usecases/run-graph"
	"github.com/Gadzet005/shortcut/pkg/app/di"
	"github.com/Gadzet005/shortcut/pkg/app/lifecycle"
	"github.com/Gadzet005/shortcut/pkg/errors"
	httpmiddleware "github.com/Gadzet005/shortcut/pkg/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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

	tracingEnabled := s.Config().TracingConfig.Enabled
	var traceRepo trace.Repo

	if tracingEnabled {
		mongoCfg := s.Config().MongoConfig
		mongoClient, err := mongo.Connect(options.Client().ApplyURI(mongoCfg.URI))
		if err != nil {
			return errors.WrapFailf(err, "failed to connect to mongodb")
		}
		c.AddStopper(func(ctx context.Context) error {
			return mongoClient.Disconnect(ctx)
		})

		db := mongoClient.Database(mongoCfg.Database)
		traceRepo, err = tracemongo.NewMongoRepo(c.Context(), db)
		if err != nil {
			return errors.WrapFailf(err, "failed to create trace repo")
		}
	}

	postgresCfg := s.Config().PostgresConfig

	postgresDB, err := sqlx.Connect("postgres", postgresCfg.URI)
	if err != nil {
		return errors.WrapFail(err, "failed to create postgres repo")
	}

	defer postgresDB.Close()

	failuresRepo, err := failurepostgres.NewPostgresRepo(postgresDB)

	r := s.HTTP("shortcut")
	r.Use(
		httpmiddleware.RequestID(),
		httpmiddleware.ZapLogger(s.Logger()),
		httpmiddleware.Metrics("shortcut"),
		httpmiddleware.Recover(),
	)

	runGraphUC := rungraph.NewUseCase(client, s.Logger(), localRepo, traceRepo)
	revertRequestUC := revertrequest.NewUseCase(client, s.Logger(), localRepo, traceRepo, failuresRepo)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.Static("/ui", "./web/dist")

	handlerBase := graphhandler.NewHandlerBase(runGraphUC, revertRequestUC, tracingEnabled)
	r.Any("run/:namespace_id/*path", handlerBase.RunGraph)
	r.Any("run/:request_id/:revert_strategy", handlerBase.)

	if tracingEnabled {
		traceHandlerBase := tracehandler.NewHandlerBase(traceRepo)
		r.GET("/trace/:request_id", traceHandlerBase.GetTrace)
	}

	return nil
}
