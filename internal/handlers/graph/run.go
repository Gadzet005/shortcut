package graphhandler

import (
	"net/http"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	"github.com/Gadzet005/shortcut/internal/domain/trace"
	"github.com/Gadzet005/shortcut/pkg/errors"
	httpcontext "github.com/Gadzet005/shortcut/pkg/http/context"
	httpmiddleware "github.com/Gadzet005/shortcut/pkg/http/middleware"
	shortcutapi "github.com/Gadzet005/shortcut/pkg/shortcut/api"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h handlerBase) RunGraph(c *gin.Context) {
	logger := httpcontext.GetLogger(c).Named("RunGraph")

	namespaceID := c.Param("namespace_id")
	if namespaceID == "" {
		logger.Warn("namespace_id is required")
		c.JSON(http.StatusBadRequest, errors.Error("namespace_id is required"))
		return
	}

	logger = logger.With(zap.String("namespace_id", namespaceID))

	data, err := c.GetRawData()
	if err != nil {
		logger.Warn("failed to read request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, errors.Error("failed to read request body"))
		return
	}

	httpRequest := shortcutapi.HttpRequest{
		Method:  c.Request.Method,
		Path:    c.Param("path"),
		Headers: c.Request.Header,
		Query:   c.Request.URL.Query(),
		Body:    data,
	}

	ctx := c.Request.Context()
	if h.tracingEnabled {
		requestID, _ := c.Get(httpmiddleware.RequestIDKey)
		collector := trace.NewCollector(trace.RequestID(requestID.(string)))
		ctx = trace.WithCollector(ctx, collector)
	}

	resp, err := h.runGraphUC.RunGraph(ctx, graph.NamespaceID(namespaceID), httpRequest)
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

	for header, values := range resp.Headers {
		for _, value := range values {
			c.Header(header, value)
		}
	}

	contentType := ""
	if ct, ok := resp.Headers["Content-Type"]; ok && len(ct) > 0 {
		contentType = ct[0]
	}
	c.Data(resp.StatusCode, contentType, []byte(resp.Body))
}
