package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/Gadzet005/shortcut/demo/services/product-service/handlers"
	"github.com/Gadzet005/shortcut/pkg/app/di"
	"github.com/Gadzet005/shortcut/pkg/app/lifecycle"
	"github.com/Gadzet005/shortcut/pkg/errors"
	"github.com/Gadzet005/shortcut/pkg/shortcut"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func newService() lifecycle.App {
	s := &service{}
	s.Container = di.NewContainer[Config](s)
	return di.NewApp(s.Container)
}

type service struct {
	*di.Container[Config]
}

func (s *service) Name() string {
	return "product-service"
}

func (s *service) Run(ctx lifecycle.Context) error {
	pool, err := pgxpool.New(ctx.Context(), s.Config().Postgres.URL)
	if err != nil {
		return errors.WrapFail(err, "connect postgres")
	}
	ctx.AddStopper(func(_ context.Context) error {
		pool.Close()
		return nil
	})

	if err := handlers.InitSchema(ctx.Context(), pool); err != nil {
		return errors.WrapFail(err, "init schema")
	}
	handlers.SetPool(pool)

	products, err := handlers.LoadProducts(filepath.Join(mockDataDir(), "products.yaml"))
	if err != nil {
		return errors.WrapFail(err, "load mock products")
	}
	if err := handlers.SeedProducts(ctx.Context(), pool, products); err != nil {
		return errors.WrapFail(err, "seed products")
	}

	r := s.HTTP("product-service")
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	r.POST("/get-product", shortcut.New(handlers.GetProduct, s.Logger()))
	r.POST("/list-products", shortcut.New(handlers.ListProducts, s.Logger()))
	r.POST("/validate-product", shortcut.New(handlers.GetProduct, s.Logger()))

	return nil
}

func mockDataDir() string {
	if dir := os.Getenv("MOCK_DATA_DIR"); dir != "" {
		return dir
	}
	return "/app/mock-data"
}
