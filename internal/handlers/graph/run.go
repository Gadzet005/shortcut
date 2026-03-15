package graphhandler

import (
	"encoding/json"
	"net/http"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	rungraph "github.com/Gadzet005/shortcut/internal/usecases/run-graph"
	"github.com/Gadzet005/shortcut/pkg/errors"
	httpcontext "github.com/Gadzet005/shortcut/pkg/http/context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h handlerBase) RunGraph(c *gin.Context) {
	logger := httpcontext.GetLogger(c).Named("RunGraph")

	graphID := c.Param("graph_id")
	if graphID == "" {
		logger.Warn("graph_id is required")
		c.JSON(http.StatusBadRequest, errors.Error("graph_id is required"))
		return
	}

	logger = logger.With(zap.String("graph_id", graphID))

	data, err := c.GetRawData()
	if err != nil {
		logger.Warn("failed to read request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, errors.Error("failed to read request body"))
		return
	}

	resp, err := h.runGraphUC.RunGraph(c.Request.Context(), rungraph.Request{
		GraphID: graph.ID(graphID),
		Data:    data,
	})
	switch {
	case errors.Is(err, graph.ErrNotFound):
		logger.Warn("graph not found", zap.Error(err))
		c.JSON(http.StatusNotFound, errors.Error("graph not found"))
		return
	case err != nil:
		logger.Error("failed to run graph", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errors.Error("failed to run graph"))
		return
	}

	c.JSON(http.StatusOK, json.RawMessage(resp.Data))
}
