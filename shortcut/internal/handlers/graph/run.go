package graphhandler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Gadzet005/shortcut/shortcut/internal/domain/graph"
	rungraph "github.com/Gadzet005/shortcut/shortcut/internal/usecases/run-graph"
	errorsutils "github.com/Gadzet005/shortcut/shortcut/pkg/utils/errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h handlerBase) RunGraph(c *gin.Context) {
	logger := h.logger.Named("RunGraph")

	graphID := c.Request.PathValue("graph_id")
	if graphID == "" {
		logger.Warn("graph_id is required")
		c.JSON(http.StatusBadRequest, errorsutils.NewJSONError("graph_id is required"))
		return
	}

	logger = h.logger.With(zap.String("graph_id", graphID))

	data, err := c.GetRawData()
	if err != nil {
		logger.Warn("failed to read request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, errorsutils.NewJSONError("failed to read request body"))
		return
	}

	resp, err := h.runGraphUC.RunGraph(c.Request.Context(), rungraph.RunGraphRequest{
		GraphID: graph.GraphID(graphID),
		Data:    data,
	})
	switch {
	case errors.Is(err, graph.ErrNotFound):
		logger.Warn("graph not found", zap.Error(err))
		c.JSON(http.StatusNotFound, errorsutils.NewJSONError("graph not found"))
		return
	case err != nil:
		logger.Error("failed to run graph", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errorsutils.NewJSONError("failed to run graph"))
		return
	}

	c.JSON(http.StatusOK, json.RawMessage(resp.Data))
}
