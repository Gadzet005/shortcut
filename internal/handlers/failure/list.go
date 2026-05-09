package failurehandler

import (
	"net/http"

	"github.com/Gadzet005/shortcut/pkg/errors"
	httpcontext "github.com/Gadzet005/shortcut/pkg/http/context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h handlerBase) List(c *gin.Context) {
	logger := httpcontext.GetLogger(c).Named("ListFailures")

	namespaceID := c.Param("namespace_id")
	graphID := c.Param("graph_id")
	if namespaceID == "" || graphID == "" {
		c.JSON(http.StatusBadRequest, errors.Error("namespace_id and graph_id are required"))
		return
	}

	failures, err := h.failureRepo.ListByGraph(c.Request.Context(), namespaceID, graphID)
	if err != nil {
		logger.Error("list failures failed",
			zap.String("namespace_id", namespaceID),
			zap.String("graph_id", graphID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, errors.Error("list failures failed"))
		return
	}

	resp := make([]failureResponse, len(failures))
	for i, f := range failures {
		resp[i] = toResponse(f)
	}
	c.JSON(http.StatusOK, resp)
}
