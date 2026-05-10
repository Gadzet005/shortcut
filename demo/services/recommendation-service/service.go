package main

import (
	"os"
	"path/filepath"

	"github.com/Gadzet005/shortcut/demo/services/recommendation-service/handlers"
	"github.com/Gadzet005/shortcut/pkg/app/di"
	"github.com/Gadzet005/shortcut/pkg/app/lifecycle"
	"github.com/Gadzet005/shortcut/pkg/errors"
	"github.com/Gadzet005/shortcut/pkg/shortcut"
	"github.com/gin-gonic/gin"
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
	return "recommendation-service"
}

func (s *service) Run(_ lifecycle.Context) error {
	entries, err := handlers.LoadCatalog(filepath.Join(mockDataDir(), "products.yaml"))
	if err != nil {
		return errors.WrapFail(err, "load mock catalog")
	}
	handlers.SetCatalog(entries)

	r := s.HTTP("recommendation-service")
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	r.POST("/get-recommendations", shortcut.New(handlers.GetRecommendations, s.Logger()))

	return nil
}

func mockDataDir() string {
	if dir := os.Getenv("MOCK_DATA_DIR"); dir != "" {
		return dir
	}
	return "/app/mock-data"
}
