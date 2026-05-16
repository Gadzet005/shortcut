package httpmiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

func newTestMiddleware(t *testing.T) (gin.HandlerFunc, *httpServiceMetrics) {
	t.Helper()
	reg := prometheus.NewPedanticRegistry()
	m := newHTTPMetrics("test", promauto.With(reg))
	return metricsHandler(m), m
}

func doRequest(t *testing.T, r *gin.Engine, method, path string) int {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestMetrics_RecordsStatusCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, m := newTestMiddleware(t)

	r := gin.New()
	r.Use(handler)
	r.GET("/ok", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/fail", func(c *gin.Context) { c.Status(http.StatusInternalServerError) })
	r.GET("/notfound", func(c *gin.Context) { c.Status(http.StatusNotFound) })

	require.Equal(t, http.StatusOK, doRequest(t, r, http.MethodGet, "/ok"))
	require.Equal(t, http.StatusOK, doRequest(t, r, http.MethodGet, "/ok"))
	require.Equal(t, http.StatusInternalServerError, doRequest(t, r, http.MethodGet, "/fail"))
	require.Equal(t, http.StatusNotFound, doRequest(t, r, http.MethodGet, "/notfound"))

	require.Equal(t, 2.0, testutil.ToFloat64(m.codesTotal.WithLabelValues(http.MethodGet, "/ok", "200")))
	require.Equal(t, 1.0, testutil.ToFloat64(m.codesTotal.WithLabelValues(http.MethodGet, "/fail", "500")))
	require.Equal(t, 1.0, testutil.ToFloat64(m.codesTotal.WithLabelValues(http.MethodGet, "/notfound", "404")))

	require.Equal(t, 2.0, testutil.ToFloat64(m.requestsCnt.WithLabelValues(http.MethodGet, "/ok")))
}

func TestMetrics_UnmatchedRouteUsesUnknownLabel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, m := newTestMiddleware(t)

	r := gin.New()
	r.Use(handler)

	require.Equal(t, http.StatusNotFound, doRequest(t, r, http.MethodGet, "/missing"))

	require.Equal(t, 1.0, testutil.ToFloat64(m.codesTotal.WithLabelValues(http.MethodGet, defaultEndpointName, "404")))
	require.Equal(t, 1.0, testutil.ToFloat64(m.requestsCnt.WithLabelValues(http.MethodGet, defaultEndpointName)))
}
