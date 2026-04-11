package tracehandler

import (
	"net/http"
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/trace"
	"github.com/Gadzet005/shortcut/pkg/errors"
	httpcontext "github.com/Gadzet005/shortcut/pkg/http/context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h handlerBase) GetTrace(c *gin.Context) {
	logger := httpcontext.GetLogger(c).Named("GetTrace")

	requestID := c.Param("request_id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request_id is required"})
		return
	}

	t, err := h.traceRepo.GetByRequestID(c.Request.Context(), trace.RequestID(requestID))
	switch {
	case errors.Is(err, trace.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "trace not found"})
		return
	case err != nil:
		logger.Error("failed to get trace", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, toResponse(t))
}

func toResponse(t trace.Trace) traceResponse {
	nodeTraces := make([]nodeTraceResponse, len(t.NodeTraces))
	for i, nt := range t.NodeTraces {
		nodeTraces[i] = nodeTraceResponse{
			NodeID:     nt.NodeID,
			StartedAt:  nt.StartedAt.Format(time.RFC3339Nano),
			FinishedAt: nt.FinishedAt.Format(time.RFC3339Nano),
			DurationMs: nt.DurationMs,
			StatusCode: nt.StatusCode,
			RetryCount: nt.RetryCount,
			Error:      nt.Error,
		}
	}
	return traceResponse{
		RequestID:   t.RequestID.String(),
		NamespaceID: t.NamespaceID,
		GraphID:     t.GraphID,
		Method:      t.Method,
		Path:        t.Path,
		StartedAt:   t.StartedAt.Format(time.RFC3339Nano),
		FinishedAt:  t.FinishedAt.Format(time.RFC3339Nano),
		DurationMs:  t.DurationMs,
		Status:      t.Status.String(),
		Error:       t.Error,
		NodeTraces:  nodeTraces,
	}
}
