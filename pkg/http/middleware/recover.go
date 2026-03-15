package httpmiddleware

import (
	httpcontext "github.com/Gadzet005/shortcut/pkg/http/context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			logger := httpcontext.GetLogger(c)

			if err := recover(); err != nil {
				logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.Stack("stack"),
				)
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}
