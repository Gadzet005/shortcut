package metrics

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
	
	"go.uber.org/zap"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewService(port int, logger *zap.Logger) (service, error) {
	return service{
		logger: logger,
		srv: &http.Server{
			Addr:              fmt.Sprintf(":%d", port),
			Handler:           promhttp.Handler(),
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