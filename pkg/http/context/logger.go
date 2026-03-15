package httpcontext

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type loggerKeyType struct{}

var loggerKey = loggerKeyType{}

func SetLogger(c *gin.Context, logger *zap.Logger) {
	c.Set(loggerKey, logger)
}

func GetLogger(c *gin.Context) *zap.Logger {
	logger, ok := c.Get(loggerKey)
	if !ok {
		panic("logger not found in context")
	}
	return logger.(*zap.Logger)
}
