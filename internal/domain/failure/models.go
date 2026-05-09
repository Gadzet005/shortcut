package failure

import (
	"errors"
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/trace"
)

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusDone       Status = "done"
	StatusFailed     Status = "failed"
)

var ErrNotFound = errors.New("failure not found")

type Status string

func (s Status) String() string {
	return string(s)
}

type Failure struct {
	RequestID      string
	NamespaceID    string
	GraphID        string
	Method         string
	Path           string
	StartedAt      time.Time
	FinishedAt     time.Time
	DurationMs     int64
	Status         Status
	Error          string
	NodeTraces     []trace.NodeTrace
	ReadyToRetryAt time.Time
	NumRetry       int64
	RequestBody    []byte
}

func (f Failure) VisitedNodes() []string {
	visited := make([]string, 0, len(f.NodeTraces))
	for _, nt := range f.NodeTraces {
		if nt.Error == "" {
			visited = append(visited, nt.NodeID)
		}
	}
	return visited
}
