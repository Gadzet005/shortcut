package trace

import "sync"

type Collector struct {
	mu         sync.Mutex
	requestID  RequestID
	nodeTraces []NodeTrace
}

func NewCollector(requestID RequestID) *Collector {
	return &Collector{
		requestID:  requestID,
		nodeTraces: make([]NodeTrace, 0),
	}
}

func (c *Collector) RequestID() RequestID {
	return c.requestID
}

func (c *Collector) Add(nt NodeTrace) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nodeTraces = append(c.nodeTraces, nt)
}

func (c *Collector) NodeTraces() []NodeTrace {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]NodeTrace, len(c.nodeTraces))
	copy(result, c.nodeTraces)
	return result
}
