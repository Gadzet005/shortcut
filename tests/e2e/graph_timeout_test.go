package e2e

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Graph-level timeout test.
// The graph has two sequential nodes, each sleeping 100ms (chain = ~200ms total).
// A graph-level timeout-ms: 150 is configured, so the second node is cancelled
// by the context deadline before it finishes.

func TestGraphTimeout(t *testing.T) {
	start := time.Now()

	resp, err := http.Get(shortcutURL + "/run/timeout-test/timeout-test/graph-timeout")
	require.NoError(t, err)
	defer resp.Body.Close()

	elapsed := time.Since(start)

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Less(t, elapsed, 400*time.Millisecond, "graph should have been cut short by the graph-level timeout (150ms), not run the full chain (~200ms + overhead)")
}
