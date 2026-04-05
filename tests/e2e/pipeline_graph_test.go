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
		name  string
		check func(t *testing.T, resp longChainResponse)
	}{
		{
			name: "value passes through all five nodes correctly",
			check: func(t *testing.T, resp longChainResponse) {
				require.Equal(t, http.StatusOK, resp.StatusCode)
				require.Equal(t, 50, resp.Result)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := runLongChain(t)
			tc.check(t, resp)
		})
	}
}

type longChainResponse struct {
	StatusCode int
	Result     int
}

func runLongChain(t *testing.T) longChainResponse {
	t.Helper()

	resp, err := http.Get(shortcutURL + "/run/pipeline/pipeline/long-chain")
	require.NoError(t, err)
	defer resp.Body.Close()

	result := longChainResponse{StatusCode: resp.StatusCode}

	if resp.StatusCode == http.StatusOK {
		require.Contains(t, resp.Header.Get("Content-Type"), "application/json")
		body := make(map[string]int)
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		result.Result = body["result"]
	}

	return result
}
