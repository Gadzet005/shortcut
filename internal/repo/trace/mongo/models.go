package tracemongo

import (
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/trace"
)

type traceDocument struct {
	RequestID   string              `bson:"request_id"`
	NamespaceID string              `bson:"namespace_id"`
	GraphID     string              `bson:"graph_id"`
	Method      string              `bson:"method"`
	Path        string              `bson:"path"`
	StartedAt   time.Time           `bson:"started_at"`
	FinishedAt  time.Time           `bson:"finished_at"`
	DurationMs  int64               `bson:"duration_ms"`
	Status      string              `bson:"status"`
	Error       string              `bson:"error,omitempty"`
	NodeTraces  []nodeTraceDocument `bson:"node_traces"`
}

type nodeTraceDocument struct {
	NodeID     string    `bson:"node_id"`
	StartedAt  time.Time `bson:"started_at"`
	FinishedAt time.Time `bson:"finished_at"`
	DurationMs int64     `bson:"duration_ms"`
	StatusCode int       `bson:"status_code,omitempty"`
	RetryCount int       `bson:"retry_count,omitempty"`
	Error      string    `bson:"error,omitempty"`
}

func toDocument(t trace.Trace) traceDocument {
	nodeTraces := make([]nodeTraceDocument, len(t.NodeTraces))
	for i, nt := range t.NodeTraces {
		nodeTraces[i] = nodeTraceDocument{
			NodeID:     nt.NodeID,
			StartedAt:  nt.StartedAt,
			FinishedAt: nt.FinishedAt,
			DurationMs: nt.DurationMs,
			StatusCode: nt.StatusCode,
			RetryCount: nt.RetryCount,
			Error:      nt.Error,
		}
	}
	return traceDocument{
		RequestID:   t.RequestID.String(),
		NamespaceID: t.NamespaceID,
		GraphID:     t.GraphID,
		Method:      t.Method,
		Path:        t.Path,
		StartedAt:   t.StartedAt,
		FinishedAt:  t.FinishedAt,
		DurationMs:  t.DurationMs,
		Status:      t.Status.String(),
		Error:       t.Error,
		NodeTraces:  nodeTraces,
	}
}

func fromDocument(d traceDocument) trace.Trace {
	nodeTraces := make([]trace.NodeTrace, len(d.NodeTraces))
	for i, nt := range d.NodeTraces {
		nodeTraces[i] = trace.NodeTrace{
			NodeID:     nt.NodeID,
			StartedAt:  nt.StartedAt,
			FinishedAt: nt.FinishedAt,
			DurationMs: nt.DurationMs,
			StatusCode: nt.StatusCode,
			RetryCount: nt.RetryCount,
			Error:      nt.Error,
		}
	}
	return trace.Trace{
		RequestID:   trace.RequestID(d.RequestID),
		NamespaceID: d.NamespaceID,
		GraphID:     d.GraphID,
		Method:      d.Method,
		Path:        d.Path,
		StartedAt:   d.StartedAt,
		FinishedAt:  d.FinishedAt,
		DurationMs:  d.DurationMs,
		Status:      trace.TraceStatus(d.Status),
		Error:       d.Error,
		NodeTraces:  nodeTraces,
	}
}
