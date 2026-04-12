package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// Graph: input → call-echo (http-adapter → mock-service GET /node-rwr-test/echo)
// The mock handler returns the query keys it received, so we can assert
// that node-rwr is not leaked to downstream nodes.

func TestNodeRwr_InvalidFormat(t *testing.T) {
	testCases := []struct {
		name    string
		nodeRwr string
	}{
		{name: "only one part", nodeRwr: "call-echo"},
		{name: "only two parts", nodeRwr: "call-echo:localhost"},
		{name: "empty node name", nodeRwr: ":localhost:9090"},
		{name: "empty host", nodeRwr: "call-echo::9090"},
		{name: "empty port", nodeRwr: "call-echo:localhost:"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(shortcutURL + "/run/node-rwr/echo?node-rwr=" + tc.nodeRwr)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestNodeRwr_UnknownNode(t *testing.T) {
	resp, err := http.Get(shortcutURL + "/run/node-rwr/echo?node-rwr=nonexistent:localhost:9090")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var body map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.Contains(t, body["error"], "nonexistent")
}

func TestNodeRwr_OverrideApplied(t *testing.T) {
	t.Run("baseline without override succeeds", func(t *testing.T) {
		resp, err := http.Get(shortcutURL + "/run/node-rwr/echo")
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("override to unreachable host fails", func(t *testing.T) {
		// 127.0.0.1:9999 is connection-refused inside the container — fast failure.
		// If the override were not applied, the request would go to mock-service and succeed.
		resp, err := http.Get(shortcutURL + "/run/node-rwr/echo?node-rwr=call-echo:127.0.0.1:9999")
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

// Graph redirect: input → call-echo (http-adapter → 127.0.0.1:9999 — unreachable by default)
// node-rwr redirects call-echo to mock-service:9001, making the request succeed.
func TestNodeRwr_SuccessfulOverride(t *testing.T) {
	t.Run("without override fails — endpoint unreachable", func(t *testing.T) {
		resp, err := http.Get(shortcutURL + "/run/node-rwr/redirect")
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("with override redirects to reachable host and succeeds", func(t *testing.T) {
		resp, err := http.Get(shortcutURL + "/run/node-rwr/redirect?node-rwr=call-echo:mock-service:9001")
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var body struct {
			QueryKeys []string `json:"query_keys"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		require.NotNil(t, body.QueryKeys)
	})
}

func TestNodeRwr_NotLeakedToDownstream(t *testing.T) {
	// Pass node-rwr alongside a regular query param.
	// The override points back to the same mock-service so the request succeeds.
	// We then assert that mock-service did NOT receive node-rwr in the query.
	resp, err := http.Get(shortcutURL + "/run/node-rwr/echo?name=world&node-rwr=call-echo:mock-service:9001")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body struct {
		QueryKeys []string `json:"query_keys"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.Contains(t, body.QueryKeys, "name")
	require.NotContains(t, body.QueryKeys, "node-rwr")
}
