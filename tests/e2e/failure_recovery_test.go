package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestFailureRevertStrategy_CallsRevertOnVisitedNodes verifies that when a graph
// fails with the synchronous `revert` strategy, the engine walks the visited
// nodes in reverse topological order and calls each node's revert endpoint.
//
// Graph layout (failure-strategy: revert):
//
//	input → step-a (succeeds, has revert-path) → step-b (always fails 500)
//
// Expected: step-a's revert endpoint is hit exactly once, keyed by the failed
// request's X-Request-Id. step-b is never reverted (it never succeeded).
func TestFailureRevertStrategy_CallsRevertOnVisitedNodes(t *testing.T) {
	sessionID := sessionID(t)

	url := fmt.Sprintf("%s/run/failure-test/failure-test/revert-demo?session_id=%s",
		shortcutURL, sessionID)
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode,
		"graph must fail because step-b returns 400")

	requestID := resp.Header.Get("X-Request-Id")
	require.NotEmpty(t, requestID, "X-Request-Id header must be set on the response")

	require.Equal(t, 1, queryFailureCount(t, sessionID+"/step-a-runs"),
		"step-a must run exactly once during the original execution")
	require.Equal(t, 1, queryFailureCount(t, sessionID+"/step-b-runs"),
		"step-b must run exactly once and fail")
	require.Equal(t, 1, queryFailureCount(t, requestID+"/step-a-reverts"),
		"step-a's revert endpoint must be called exactly once for this request")
}

// TestFailureFinishStrategy_RetriesFailedNodes verifies that when a graph fails
// with the synchronous `finish` strategy, the engine re-runs the unvisited
// (failed) nodes via TryFinish — visited nodes are skipped.
//
// Graph layout (failure-strategy: finish):
//
//	input → flaky-final (fails on call 1, succeeds on call 2)
//
// On the original Run flaky-final fails (call 1) and the client receives 500.
// The finish strategy then synchronously re-runs only flaky-final (input was
// visited and is skipped), bumping its call counter to 2.
func TestFailureFinishStrategy_RetriesFailedNodes(t *testing.T) {
	sessionID := sessionID(t)

	url := fmt.Sprintf("%s/run/failure-test/failure-test/finish-demo?session_id=%s&fail_count=1",
		shortcutURL, sessionID)
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode,
		"client gets the original failure response; finish only runs side-effects")

	require.Equal(t, 2, queryFailureCount(t, sessionID+"/flaky-final-calls"),
		"flaky-final must run twice: once during the original failure and once via finish")
}

// queryFailureCount asks mock-service (through the shortcut graph) for the
// counter stored under the given key.
func queryFailureCount(t *testing.T, key string) int {
	t.Helper()
	url := fmt.Sprintf("%s/run/failure-test/failure-test/stats?key=%s", shortcutURL, key)
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body struct {
		Count int `json:"count"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	return body.Count
}
