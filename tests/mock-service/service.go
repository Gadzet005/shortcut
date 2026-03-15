package main

import (
	"github.com/Gadzet005/shortcut/pkg/app/di"
	"github.com/Gadzet005/shortcut/pkg/app/lifecycle"
	"github.com/Gadzet005/shortcut/pkg/shortcut"
	"github.com/gin-gonic/gin"
)

func newService() lifecycle.App {
	s := &service{}
	s.Container = di.NewContainer[any](s)
	return di.NewApp(s.Container)
}

type service struct {
	*di.Container[any]
}

func (s *service) Name() string {
	return "mock-service"
}

func (s *service) Run(ctx lifecycle.Context) error {
	r := s.HTTP("mock-service")
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.POST("/echo", shortcut.New(echoHandler, s.Logger()))
	r.POST("/sum", shortcut.New(sumHandler, s.Logger()))
	return nil
}
