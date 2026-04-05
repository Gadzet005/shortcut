package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Retry tests exercise the exponential-backoff retry logic in defaultNodeExecutor.
//
// The mock /retry-test/flaky endpoint accepts:
//   - session_id  — isolates per-test call counters
//   - fail_count  — how many times to return an error before succeeding
//   - fail_status — which HTTP status to return while failing (default 500)
//
// Two graphs are configured in the retry-test namespace:
//   - recoverable      — endpoint with retries-num: 2  (3 total attempts)
//   - exhausted-retries — endpoint with retries-num: 1  (2 total attempts)

func TestRetrySucceeds(t *testing.T) {
	// fail_count=2 with retries_num=2: fails on attempts 1 & 2, succeeds on attempt 3.
	sessionID := sessionID(t)
	resp := callRetryGraph(t, "recoverable", sessionID, "2", "")

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, 3, resp.TotalAttempts, "expected 3 attempts: 2 failures + 1 success")
}

func TestRetryExhausted(t *testing.T) {
	// fail_count=999 with retries_num=1: fails on attempts 1 & 2, no more retries left.
	sessionID := sessionID(t)
	resp := callRetryGraph(t, "exhausted-retries", sessionID, "999", "")

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestNoRetryOn4xx(t *testing.T) {
	// fail_status=400 with retries_num=2: the first call returns 400 (bad request).
	// 4xx errors must NOT be retried — the graph must return 400 immediately.
	// If the executor incorrectly retried, the second call would return 200 (fail_count=1).
	sessionID := sessionID(t)
	resp := callRetryGraph(t, "recoverable", sessionID, "1", "400")

	require.Equal(t, http.StatusBadRequest, resp.StatusCode,
		"4xx errors must not be retried: expected 400, not 200")
}

func TestRetryExponentialBackoffTiming(t *testing.T) {
	// fail_count=2, retries_num=2, initial_interval=10ms, multiplier=2 → backoffs: 10ms, 20ms.
	// Total overhead: ≥30ms. We verify the graph takes at least that long, confirming
	// backoff actually happens (not an instant retry loop).
	sessionID := sessionID(t)

	start := time.Now()
	resp := callRetryGraph(t, "recoverable", sessionID, "2", "")
	elapsed := time.Since(start)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.GreaterOrEqual(t, elapsed, 30*time.Millisecond,
		"expected at least 30ms of backoff (10ms + 20ms) for 2 retries")
}

// ── helpers ───────────────────────────────────────────────────────────────────

type retryGraphResponse struct {
	StatusCode    int
	TotalAttempts int
}

// callRetryGraph calls GET /run/retry-test/retry-test/{graph} with the given params.
// failStatus is optional; pass "" to use the mock's default (500).
func callRetryGraph(t *testing.T, graph, sessionID, failCount, failStatus string) retryGraphResponse {
	t.Helper()

	url := fmt.Sprintf("%s/run/retry-test/retry-test/%s?session_id=%s&fail_count=%s",
		shortcutURL, graph, sessionID, failCount)
	if failStatus != "" {
		url += "&fail_status=" + failStatus
	}

	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	result := retryGraphResponse{StatusCode: resp.StatusCode}
	if resp.StatusCode == http.StatusOK {
		body := struct {
			TotalAttempts int `json:"total_attempts"`
		}{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		result.TotalAttempts = body.TotalAttempts
	}
	return result
}

// sessionID returns a unique string per test to isolate the mock's call counters.
func sessionID(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("%s-%d", t.Name(), time.Now().UnixNano())
}
