package e2e

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestNodeCaching checks that a node result is served from cache within the TTL
// and recomputed after the TTL expires.
//
// Graph: input → counter (cache ttl=1s) → output
// The counter mock endpoint increments a global counter on every real call.
// Within the TTL both requests must return the same count.
// After waiting for the TTL to expire the next request must return a higher count.
func TestNodeCaching(t *testing.T) {
	first := getCounter(t)
	require.Equal(t, http.StatusOK, first.StatusCode)

	second := getCounter(t)
	require.Equal(t, http.StatusOK, second.StatusCode)
	require.Equal(t, first.Count, second.Count, "second request should return cached result")

	// Wait for TTL to expire (1s + small buffer).
	time.Sleep(1200 * time.Millisecond)

	third := getCounter(t)
	require.Equal(t, http.StatusOK, third.StatusCode)
	require.Greater(t, third.Count, first.Count, "count must increase after cache expires")
}

// TestNodeCachingIndependent checks that cache entries are isolated per unique
// request: a cached result for one set of query params must not interfere with
// the cache entry for a different set of query params.
//
// Graph: input → counter (cache ttl=1s) → output
// Sequence:
//  1. key=a  → real call (count N)
//  2. key=a  → cached   (count N)
//  3. key=b  → real call (count N+1, separate cache entry)
//  4. key=a  → still cached (count N, unaffected by the key=b call)
//  5. key=b  → cached   (count N+1)
func TestNodeCachingIndependent(t *testing.T) {
	a1 := getCounterWithKey(t, "a")
	require.Equal(t, http.StatusOK, a1.StatusCode)

	a2 := getCounterWithKey(t, "a")
	require.Equal(t, http.StatusOK, a2.StatusCode)
	require.Equal(t, a1.Count, a2.Count, "key=a: second request should be served from cache")

	b1 := getCounterWithKey(t, "b")
	require.Equal(t, http.StatusOK, b1.StatusCode)
	require.Greater(t, b1.Count, a1.Count, "key=b: should trigger a real call with a higher count")

	a3 := getCounterWithKey(t, "a")
	require.Equal(t, http.StatusOK, a3.StatusCode)
	require.Equal(t, a1.Count, a3.Count, "key=a cache must not be affected by the key=b call")

	b2 := getCounterWithKey(t, "b")
	require.Equal(t, http.StatusOK, b2.StatusCode)
	require.Equal(t, b1.Count, b2.Count, "key=b: second request should be served from its own cache")
}

type counterResponse struct {
	StatusCode int
	Count      int64
}

func getCounter(t *testing.T) counterResponse {
	t.Helper()
	return getCounterWithKey(t, "")
}

func getCounterWithKey(t *testing.T, key string) counterResponse {
	t.Helper()

	url := shortcutURL + "/run/cache-test/cache-test/counter"
	if key != "" {
		url += "?key=" + key
	}

	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	result := counterResponse{StatusCode: resp.StatusCode}
	if resp.StatusCode == http.StatusOK {
		var body map[string]int64
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		result.Count = body["count"]
	}
	return result
}
