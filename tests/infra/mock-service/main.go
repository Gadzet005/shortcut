package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Gadzet005/shortcut/shortcut/pkg/lifecycle"
	"github.com/Gadzet005/shortcut/shortcut/pkg/shortcut"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	port := flag.Int("port", 8080, "Port to listen on")
	flag.Parse()

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	r := gin.New()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.POST("/echo", shortcut.New(echoHandler, logger))
	r.POST("/sum", shortcut.New(sumHandler, logger))

	addr := fmt.Sprintf(":%d", *port)
	service := newService(r, addr, logger)

	if err := lifecycle.Run(service); err != nil {
		logger.Fatal("Failed to run service", zap.Error(err))
	}
}

func newService(handler http.Handler, addr string, logger *zap.Logger) *service {
	return &service{
		logger: logger,
		srv: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadHeaderTimeout: 10 * time.Second,
		},
	}
}

type service struct {
	logger *zap.Logger
	srv    *http.Server
}

func (s *service) Start(ctx context.Context) error {
	go func() {
		s.logger.Info("starting mock service", zap.String("addr", s.srv.Addr))
		err := s.srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("mock service stopped with error", zap.Error(err))
		}
		s.logger.Info("mock service stopped")
	}()

	return nil
}

func (s *service) Stop(ctx context.Context) error {
	s.logger.Info("stopping mock service")
	return s.srv.Shutdown(ctx)
}
