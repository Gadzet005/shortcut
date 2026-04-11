package trace

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCollector_RequestID(t *testing.T) {
	c := NewCollector("test-id")
	require.Equal(t, RequestID("test-id"), c.RequestID())
}

func TestCollector_AddAndGet(t *testing.T) {
	c := NewCollector("req-1")

	now := time.Now()
	c.Add(NodeTrace{
		NodeID:     "node-1",
		StartedAt:  now,
		FinishedAt: now.Add(10 * time.Millisecond),
		DurationMs: 10,
		StatusCode: 200,
	})
	c.Add(NodeTrace{
		NodeID:     "node-2",
		StartedAt:  now,
		FinishedAt: now.Add(20 * time.Millisecond),
		DurationMs: 20,
		StatusCode: 200,
	})

	traces := c.NodeTraces()
	require.Equal(t, 2, len(traces))
	require.Equal(t, "node-1", traces[0].NodeID)
	require.Equal(t, "node-2", traces[1].NodeID)
}

func TestCollector_NodeTracesReturnsCopy(t *testing.T) {
	c := NewCollector("req-1")
	c.Add(NodeTrace{NodeID: "node-1"})

	traces := c.NodeTraces()
	traces[0].NodeID = "modified"

	original := c.NodeTraces()
	require.Equal(t, "node-1", original[0].NodeID)
}

func TestCollector_ConcurrentAdds(t *testing.T) {
	c := NewCollector("concurrent-test")
	n := 1000

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			c.Add(NodeTrace{
				NodeID:     "node",
				DurationMs: 1,
			})
		}()
	}
	wg.Wait()

	traces := c.NodeTraces()
	require.Equal(t, n, len(traces))
}
