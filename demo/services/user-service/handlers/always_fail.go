package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AlwaysFail(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "demo: simulated downstream failure to trigger revert",
	})
}
