package failurehandler

import (
	"net/http"

	"github.com/Gadzet005/shortcut/internal/domain/failure"
	"github.com/Gadzet005/shortcut/internal/domain/graph"
	"github.com/Gadzet005/shortcut/pkg/errors"
	httpcontext "github.com/Gadzet005/shortcut/pkg/http/context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h handlerBase) Process(c *gin.Context) {
	logger := httpcontext.GetLogger(c).Named("ProcessFailure")

	requestID := c.Param("request_id")
	strategyParam := c.Param("strategy")
	if requestID == "" || strategyParam == "" {
		c.JSON(http.StatusBadRequest, errors.Error("request_id and strategy are required"))
		return
	}

	strategyName, ok := graph.ParseFailureStrategy(strategyParam)
	if !ok {
		c.JSON(http.StatusBadRequest, errors.Error("unknown strategy "+strategyParam))
		return
	}

	logger = logger.With(
		zap.String("request_id", requestID),
		zap.String("strategy", strategyName.String()),
	)

	f, err := h.failureRepo.GetByRequestID(c.Request.Context(), requestID)
	switch {
	case errors.Is(err, failure.ErrNotFound):
		c.JSON(http.StatusNotFound, errors.Error("failure not found"))
		return
	case err != nil:
		logger.Error("get failure failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errors.Error("get failure failed"))
		return
	}

	handler := h.factory.New(strategyName, nil)
	if err := handler.Handle(c.Request.Context(), f); err != nil {
		logger.Error("process failure failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errors.Error("process failed"))
		return
	}

	logger.Info("failure processed")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
