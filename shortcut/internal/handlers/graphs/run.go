package graphshandler

import "github.com/gin-gonic/gin"

func (h handlerBase) RunGraph(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "hello",
	})
}
