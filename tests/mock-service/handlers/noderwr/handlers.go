package noderwr

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
)

type echoResponse struct {
	QueryKeys []string `json:"query_keys"`
}

// Echo handles GET /node-rwr-test/echo.
// Returns the sorted list of query parameter keys it received, so tests can
// assert that internal params (e.g. node-rwr) are not leaked to downstream nodes.
func Echo(c *gin.Context) {
	keys := make([]string, 0, len(c.Request.URL.Query()))
	for k := range c.Request.URL.Query() {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	c.JSON(http.StatusOK, echoResponse{QueryKeys: keys})
}
