package main

import (
	"github.com/Gadzet005/shortcut/pkg/app/di"
	"github.com/Gadzet005/shortcut/pkg/app/lifecycle"
	"github.com/Gadzet005/shortcut/pkg/shortcut"
	"github.com/Gadzet005/shortcut/tests/mock-service/orders"
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
	return "mock-service"
}

func (s *service) Run(ctx lifecycle.Context) error {
	r := s.HTTP("mock-service")
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	{
		g := r.Group("/orders")
		g.POST("/get-users-by-ids", shortcut.New(orders.GetUsersByIDs, s.Logger()))
		g.POST("/get-top-orders", shortcut.New(orders.GetTopOrders, s.Logger()))
		g.POST("/merge-orders-and-users", shortcut.New(orders.MergeOrdersAndUsers, s.Logger()))
	}

	return nil
}
