package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// Sequential chain: input ─► step1 ─► step2 ─► step3 ─► step4 ─► step5
// Each step adds 10. Starting value is 10, so the final result is 50.

func TestLongChain(t *testing.T) {
	testCases := []struct {
		name       string
		check      func(t *testing.T, resp longChainResponse)
		checkTrace func(t *testing.T, tr traceResponse)
	}{
		{
			name: "value passes through all five nodes correctly",
			check: func(t *testing.T, resp longChainResponse) {
				require.Equal(t, http.StatusOK, resp.StatusCode)
				require.Equal(t, 50, resp.Result)
			},
			checkTrace: func(t *testing.T, tr traceResponse) {
				require.Equal(t, "ok", tr.Status)
				require.Equal(t, "pipeline", tr.NamespaceID)
				require.Equal(t, "long_chain", tr.GraphID)
				require.Equal(t, http.MethodGet, tr.Method)
				require.Len(t, tr.NodeTraces, 6)

				input := findNodeTrace(t, tr, "input")
				require.Equal(t, 0, input.StatusCode) // transparent node — no HTTP call
				require.Empty(t, input.Error)

				for _, stepName := range []string{"step1", "step2", "step3", "step4", "step5"} {
					step := findNodeTrace(t, tr, stepName)
					require.Equal(t, http.StatusOK, step.StatusCode, "node %s", stepName)
					require.Empty(t, step.Error, "node %s", stepName)
					require.GreaterOrEqual(t, step.DurationMs, int64(0))
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := runLongChain(t)
			tc.check(t, resp)
			if tc.checkTrace != nil {
				tr := getTrace(t, shortcutURL, resp.RequestID)
				tc.checkTrace(t, tr)
			}
		})
	}
}

type longChainResponse struct {
	StatusCode int
	RequestID  string
	Result     int
}

func runLongChain(t *testing.T) longChainResponse {
	t.Helper()

	resp, err := http.Get(shortcutURL + "/run/pipeline/pipeline/long-chain")
	require.NoError(t, err)
	defer resp.Body.Close()

	result := longChainResponse{
		StatusCode: resp.StatusCode,
		RequestID:  resp.Header.Get("X-Request-Id"),
	}

	if resp.StatusCode == http.StatusOK {
		require.Contains(t, resp.Header.Get("Content-Type"), "application/json")
		body := make(map[string]int)
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		result.Result = body["result"]
	}

	return result
}
