package shortcut

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HandlerFunc func(*Context) error

type Handler interface {
	Handle(*Context) error
}

func New(h HandlerFunc, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := newContext(c, logger)

		if err := ctx.parseRequest(); err != nil {
			handleError(c, err)
			return
		}

		if err := h(ctx); err != nil {
			handleError(c, err)
			return
		}
	}
}

func handleError(c *gin.Context, err error) {
	var herr *HandlerError
	if errors.As(err, &herr) {
		c.JSON(herr.StatusCode, gin.H{
			"error": herr.Message,
		})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "internal server error",
	})
}
