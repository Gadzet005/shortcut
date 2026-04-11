package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// Graph: input → call-echo (http-adapter → mock-service GET/POST /http-adapter/echo)
func TestHTTPAdapter_GET(t *testing.T) {
	testCases := []struct {
		name        string
		queryName   string
		wantStatus  int
		wantMessage string
	}{
		{
			name:        "forwards query params and returns response",
			queryName:   "world",
			wantStatus:  http.StatusOK,
			wantMessage: "hello, world",
		},
		{
			name:       "returns 400 when required param is missing",
			queryName:  "",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := shortcutURL + "/run/httpadapter/echo"
			if tc.queryName != "" {
				url += "?name=" + tc.queryName
			}

			resp, err := http.Get(url)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, tc.wantStatus, resp.StatusCode)

			if tc.wantStatus == http.StatusOK {
				var body struct {
					Message string `json:"message"`
				}
				require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
				require.Equal(t, tc.wantMessage, body.Message)

				tr := getTrace(t, shortcutURL, resp.Header.Get("X-Request-Id"))
				require.Equal(t, "ok", tr.Status)
				require.Equal(t, "httpadapter", tr.NamespaceID)
				require.Equal(t, "echo", tr.GraphID)
				require.Len(t, tr.NodeTraces, 2) // input + call-echo

				input := findNodeTrace(t, tr, "input")
				require.Equal(t, 0, input.StatusCode)
				require.Empty(t, input.Error)

				callEcho := findNodeTrace(t, tr, "call-echo")
				require.Equal(t, http.StatusOK, callEcho.StatusCode)
				require.Empty(t, callEcho.Error)
			}
		})
	}
}

func TestHTTPAdapter_POST(t *testing.T) {
	testCases := []struct {
		name        string
		body        any
		wantStatus  int
		wantMessage string
	}{
		{
			name:        "forwards request body and returns response",
			body:        map[string]string{"name": "gopher"},
			wantStatus:  http.StatusOK,
			wantMessage: "hello, gopher",
		},
		{
			name:       "returns 400 when body is missing name field",
			body:       map[string]string{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rawBody, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, shortcutURL+"/run/httpadapter/echo", bytes.NewReader(rawBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, tc.wantStatus, resp.StatusCode)

			if tc.wantStatus == http.StatusOK {
				var body struct {
					Message string `json:"message"`
				}
				require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
				require.Equal(t, tc.wantMessage, body.Message)

				tr := getTrace(t, shortcutURL, resp.Header.Get("X-Request-Id"))
				require.Equal(t, "ok", tr.Status)
				require.Equal(t, "httpadapter", tr.NamespaceID)
				require.Equal(t, "echo", tr.GraphID)
				require.Len(t, tr.NodeTraces, 2)

				callEcho := findNodeTrace(t, tr, "call-echo")
				require.Equal(t, http.StatusOK, callEcho.StatusCode)
				require.Empty(t, callEcho.Error)
			}
		})
	}
}
