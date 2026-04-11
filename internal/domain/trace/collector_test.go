package trace

import (
	"sync"
	"testing"
	"time"
)

func TestCollector_RequestID(t *testing.T) {
	c := NewCollector("test-id")
	if c.RequestID() != "test-id" {
		t.Errorf("expected request ID 'test-id', got '%s'", c.RequestID())
	}
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
	if len(traces) != 2 {
		t.Fatalf("expected 2 node traces, got %d", len(traces))
	}
	if traces[0].NodeID != "node-1" {
		t.Errorf("expected first trace node ID 'node-1', got '%s'", traces[0].NodeID)
	}
	if traces[1].NodeID != "node-2" {
		t.Errorf("expected second trace node ID 'node-2', got '%s'", traces[1].NodeID)
	}
}

func TestCollector_NodeTracesReturnsCopy(t *testing.T) {
	c := NewCollector("req-1")
	c.Add(NodeTrace{NodeID: "node-1"})

	traces := c.NodeTraces()
	traces[0].NodeID = "modified"

	original := c.NodeTraces()
	if original[0].NodeID != "node-1" {
		t.Error("NodeTraces should return a copy, but original was modified")
	}
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
	if len(traces) != n {
		t.Errorf("expected %d traces, got %d", n, len(traces))
	}
}
