package trace

import (
	"errors"
	"time"
)

const (
	TraceStatusOK    TraceStatus = "ok"
	TraceStatusError TraceStatus = "error"
)

var ErrNotFound = errors.New("trace not found")

type RequestID string

func (r RequestID) String() string {
	return string(r)
}

type Trace struct {
	RequestID   RequestID
	NamespaceID string
	GraphID     string
	Method      string
	Path        string
	StartedAt   time.Time
	FinishedAt  time.Time
	DurationMs  int64
	Status      TraceStatus
	Error       string
	NodeTraces  []NodeTrace
}

type TraceStatus string

func (t TraceStatus) String() string {
	return string(t)
}

type NodeTrace struct {
	NodeID     string
	StartedAt  time.Time
	FinishedAt time.Time
	DurationMs int64
	StatusCode int // HTTP status code from endpoint node (0 for transparent/subgraph)
	RetryCount int
	Error      string
}
