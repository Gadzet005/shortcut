package e2e

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRoutingErrors(t *testing.T) {
	testCases := []struct {
		name           string
		url            string
		method         string
		expectedStatus int
	}{
		{
			name:           "unknown namespace returns 404",
			url:            shortcutURL + "/run/nonexistent/orders/get-top-orders?limit=3",
			method:         http.MethodGet,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "unknown path in known namespace returns 404",
			url:            shortcutURL + "/run/orders/orders/nonexistent-endpoint",
			method:         http.MethodGet,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "wrong HTTP method returns 404",
			url:            shortcutURL + "/run/orders/orders/get-top-orders?limit=3",
			method:         http.MethodPost,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.url, nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func TestNodeErrors(t *testing.T) {
	testCases := []struct {
		name           string
		status         string
		expectedStatus int
	}{
		{name: "node returns 400", status: "400", expectedStatus: http.StatusBadRequest},
		{name: "node returns 401", status: "401", expectedStatus: http.StatusUnauthorized},
		{name: "node returns 403", status: "403", expectedStatus: http.StatusForbidden},
		{name: "node returns 404", status: "404", expectedStatus: http.StatusNotFound},
		{name: "node returns 500", status: "500", expectedStatus: http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(shortcutURL + "/run/test-errors/test/echo-error?status=" + tc.status)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func TestNodeBadResponse(t *testing.T) {
	t.Run("node returns non-multipart response", func(t *testing.T) {
		resp, err := http.Get(shortcutURL + "/run/test-errors/test/invalid-content-type")
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("node returns multipart without http_response item", func(t *testing.T) {
		resp, err := http.Get(shortcutURL + "/run/test-errors/test/missing-http-response")
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestNodeTimeout(t *testing.T) {
	// The slow-response endpoint sleeps for 1 second, but the node timeout is 10ms.
	// The graph must complete well before the sleep finishes.
	start := time.Now()

	resp, err := http.Get(shortcutURL + "/run/test-errors/test/slow-response")
	require.NoError(t, err)
	defer resp.Body.Close()

	elapsed := time.Since(start)

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Less(t, elapsed, 500*time.Millisecond, "graph should have timed out well before the node's 1s sleep")
}
