package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/Gadzet005/shortcut/demo/services/user-service/handlers"
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
	return "user-service"
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

	users, err := handlers.LoadUsers(filepath.Join(mockDataDir(), "users.yaml"))
	if err != nil {
		return errors.WrapFail(err, "load mock users")
	}
	if err := handlers.SeedUsers(ctx.Context(), pool, users); err != nil {
		return errors.WrapFail(err, "seed users")
	}

	r := s.HTTP("user-service")
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	r.POST("/get-user", shortcut.New(handlers.GetUser, s.Logger()))
	r.POST("/get-cart", shortcut.New(handlers.GetCart, s.Logger()))
	r.POST("/add-cart-item", shortcut.New(handlers.AddCartItem, s.Logger()))
	r.POST("/remove-cart-item", shortcut.New(handlers.RemoveCartItem, s.Logger()))
	r.POST("/revert-add-cart-item", handlers.RevertAddCartItem)
	r.POST("/revert-remove-cart-item", handlers.RevertRemoveCartItem)
	r.POST("/always-fail", handlers.AlwaysFail)

	return nil
}

func mockDataDir() string {
	if dir := os.Getenv("MOCK_DATA_DIR"); dir != "" {
		return dir
	}
	return "/app/mock-data"
}
