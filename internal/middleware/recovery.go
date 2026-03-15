package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ZapRecovery(logger *zap.Logger, stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
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
