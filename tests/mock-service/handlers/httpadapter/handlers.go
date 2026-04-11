package httpadapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type echoResponse struct {
	Message string `json:"message"`
}

// EchoGET handles GET /http-adapter/echo?name=<name>.
// Returns 400 if name query param is missing.
func EchoGET(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name query param is required"})
		return
	}
	c.JSON(http.StatusOK, echoResponse{Message: "hello, " + name})
}

// EchoPost handles POST /http-adapter/echo with JSON body {"name": "<name>"}.
// Returns 400 if name field is missing.
func EchoPost(c *gin.Context) {
	var body struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name field is required"})
		return
	}
	c.JSON(http.StatusOK, echoResponse{Message: "hello, " + body.Name})
}
