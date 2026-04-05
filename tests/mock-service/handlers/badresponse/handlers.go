package badresponse

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Gadzet005/shortcut/pkg/shortcut"
	shortcutapi "github.com/Gadzet005/shortcut/pkg/shortcut/api"
	"github.com/gin-gonic/gin"
)

// EchoError reads the desired HTTP status code from the http_request query param "status"
// and returns an error with that status code.
func EchoError(ctx *shortcut.Context) error {
	var httpRequest shortcutapi.HttpRequest
	if err := ctx.GetJSONItem("request", &httpRequest); err != nil {
		return err
	}

	statusRaw := httpRequest.Query.Get("status")
	status, err := strconv.Atoi(statusRaw)
	if err != nil {
		return shortcut.NewError(http.StatusBadRequest, "invalid status code")
	}

	return shortcut.NewError(status, "test error")
}

// InvalidContentType returns a 200 response with plain JSON instead of multipart.
func InvalidContentType(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"key": "value"})
}

// MissingHTTPResponse returns a 200 multipart response without the http_response item.
func MissingHTTPResponse(ctx *shortcut.Context) error {
	return shortcut.NewResponse().
		AddJSONItem("some_item", gin.H{"key": "value"}).
		Send(ctx)
}

// SlowResponse sleeps for 1 second before responding, simulating a slow/hanging endpoint.
func SlowResponse(c *gin.Context) {
	time.Sleep(time.Second)
	c.JSON(http.StatusOK, gin.H{"key": "value"})
}
