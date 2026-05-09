package failurehandler

import (
	"net/http"

	"github.com/Gadzet005/shortcut/internal/domain/failure"
	"github.com/Gadzet005/shortcut/pkg/errors"
	httpcontext "github.com/Gadzet005/shortcut/pkg/http/context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h handlerBase) Delete(c *gin.Context) {
	logger := httpcontext.GetLogger(c).Named("DeleteFailure")

	namespaceID := c.Param("namespace_id")
	graphID := c.Param("graph_id")
	requestID := c.Param("request_id")
	if namespaceID == "" || graphID == "" || requestID == "" {
		c.JSON(http.StatusBadRequest, errors.Error("namespace_id, graph_id and request_id are required"))
		return
	}

	existing, err := h.failureRepo.GetByRequestID(c.Request.Context(), requestID)
	switch {
	case errors.Is(err, failure.ErrNotFound):
		c.JSON(http.StatusNotFound, errors.Error("failure not found"))
		return
	case err != nil:
		logger.Error("get failure failed", zap.String("request_id", requestID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, errors.Error("delete failure failed"))
		return
	}

	if existing.NamespaceID != namespaceID || existing.GraphID != graphID {
		c.JSON(http.StatusNotFound, errors.Error("failure not found"))
		return
	}

	if err := h.failureRepo.Delete(c.Request.Context(), requestID); err != nil {
		logger.Error("delete failure failed", zap.String("request_id", requestID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, errors.Error("delete failure failed"))
		return
	}

	c.Status(http.StatusNoContent)
}
