package tracehandler

import (
	"github.com/Gadzet005/shortcut/internal/domain/trace"
)

func NewHandlerBase(traceRepo trace.Repo) handlerBase {
	return handlerBase{traceRepo: traceRepo}
}

type handlerBase struct {
	traceRepo trace.Repo
}

type traceResponse struct {
	RequestID   string              `json:"request_id"`
	NamespaceID string              `json:"namespace_id"`
	GraphID     string              `json:"graph_id"`
	Method      string              `json:"method"`
	Path        string              `json:"path"`
	StartedAt   string              `json:"started_at"`
	FinishedAt  string              `json:"finished_at"`
	DurationMs  int64               `json:"duration_ms"`
	Status      string              `json:"status"`
	Error       string              `json:"error,omitempty"`
	NodeTraces  []nodeTraceResponse `json:"node_traces"`
}

type nodeTraceResponse struct {
	NodeID       string                   `json:"node_id"`
	NodeType     string                   `json:"node_type,omitempty"`
	Dependencies []nodeDependencyResponse `json:"dependencies,omitempty"`
	StartedAt    string                   `json:"started_at"`
	FinishedAt   string                   `json:"finished_at"`
	DurationMs   int64                    `json:"duration_ms"`
	StatusCode   int                      `json:"status_code,omitempty"`
	RetryCount   int                      `json:"retry_count,omitempty"`
	Cached       bool                     `json:"cached,omitempty"`
	Error        string                   `json:"error,omitempty"`
}

type nodeDependencyResponse struct {
	NodeID string `json:"node_id"`
}
