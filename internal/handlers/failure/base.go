package failurehandler

import (
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/failure"
	"github.com/Gadzet005/shortcut/internal/domain/graph/strategy"
)

func NewHandlerBase(failureRepo failure.Repo, recovery failure.Recovery, factory strategy.Factory) handlerBase {
	return handlerBase{
		failureRepo: failureRepo,
		recovery:    recovery,
		factory:     factory,
	}
}

type handlerBase struct {
	failureRepo failure.Repo
	recovery    failure.Recovery
	factory     strategy.Factory
}

type failureResponse struct {
	RequestID      string              `json:"request_id"`
	NamespaceID    string              `json:"namespace_id"`
	GraphID        string              `json:"graph_id"`
	Method         string              `json:"method"`
	Path           string              `json:"path"`
	StartedAt      string              `json:"started_at"`
	FinishedAt     string              `json:"finished_at"`
	DurationMs     int64               `json:"duration_ms"`
	Status         string              `json:"status"`
	Error          string              `json:"error,omitempty"`
	NodeTraces     []nodeTraceResponse `json:"node_traces"`
	ReadyToRetryAt string              `json:"ready_to_retry_at"`
	NumRetry       int64               `json:"num_retry"`
}

type nodeTraceResponse struct {
	NodeID     string `json:"node_id"`
	NodeType   string `json:"node_type,omitempty"`
	StartedAt  string `json:"started_at"`
	FinishedAt string `json:"finished_at"`
	DurationMs int64  `json:"duration_ms"`
	StatusCode int    `json:"status_code,omitempty"`
	RetryCount int    `json:"retry_count,omitempty"`
	Cached     bool   `json:"cached,omitempty"`
	Error      string `json:"error,omitempty"`
}

func toResponse(f failure.Failure) failureResponse {
	nodeTraces := make([]nodeTraceResponse, len(f.NodeTraces))
	for i, nt := range f.NodeTraces {
		nodeTraces[i] = nodeTraceResponse{
			NodeID:     nt.NodeID,
			NodeType:   nt.NodeType,
			StartedAt:  nt.StartedAt.Format(time.RFC3339Nano),
			FinishedAt: nt.FinishedAt.Format(time.RFC3339Nano),
			DurationMs: nt.DurationMs,
			StatusCode: nt.StatusCode,
			RetryCount: nt.RetryCount,
			Cached:     nt.Cached,
			Error:      nt.Error,
		}
	}
	return failureResponse{
		RequestID:      f.RequestID,
		NamespaceID:    f.NamespaceID,
		GraphID:        f.GraphID,
		Method:         f.Method,
		Path:           f.Path,
		StartedAt:      f.StartedAt.Format(time.RFC3339Nano),
		FinishedAt:     f.FinishedAt.Format(time.RFC3339Nano),
		DurationMs:     f.DurationMs,
		Status:         f.Status.String(),
		Error:          f.Error,
		NodeTraces:     nodeTraces,
		ReadyToRetryAt: f.ReadyToRetryAt.Format(time.RFC3339Nano),
		NumRetry:       f.NumRetry,
	}
}
