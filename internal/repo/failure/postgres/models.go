package failurepostgres

import (
	"encoding/json"
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/failure"
	"github.com/Gadzet005/shortcut/internal/domain/trace"
)

type nodeTraceJSON struct {
	NodeID     string    `json:"node_id"`
	NodeType   string    `json:"node_type,omitempty"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	DurationMs int64     `json:"duration_ms"`
	StatusCode int       `json:"status_code,omitempty"`
	RetryCount int       `json:"retry_count,omitempty"`
	Cached     bool      `json:"cached,omitempty"`
	Error      string    `json:"error,omitempty"`
}

func encodeNodeTraces(nts []trace.NodeTrace) ([]byte, error) {
	out := make([]nodeTraceJSON, len(nts))
	for i, nt := range nts {
		out[i] = nodeTraceJSON{
			NodeID:     nt.NodeID,
			NodeType:   nt.NodeType,
			StartedAt:  nt.StartedAt,
			FinishedAt: nt.FinishedAt,
			DurationMs: nt.DurationMs,
			StatusCode: nt.StatusCode,
			RetryCount: nt.RetryCount,
			Cached:     nt.Cached,
			Error:      nt.Error,
		}
	}
	return json.Marshal(out)
}

func decodeNodeTraces(raw []byte) ([]trace.NodeTrace, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var src []nodeTraceJSON
	if err := json.Unmarshal(raw, &src); err != nil {
		return nil, err
	}
	out := make([]trace.NodeTrace, len(src))
	for i, nt := range src {
		out[i] = trace.NodeTrace{
			NodeID:     nt.NodeID,
			NodeType:   nt.NodeType,
			StartedAt:  nt.StartedAt,
			FinishedAt: nt.FinishedAt,
			DurationMs: nt.DurationMs,
			StatusCode: nt.StatusCode,
			RetryCount: nt.RetryCount,
			Cached:     nt.Cached,
			Error:      nt.Error,
		}
	}
	return out, nil
}

func scanFailure(
	requestID, namespaceID, graphID, method, path, status, errorMsg string,
	startedAt, finishedAt, readyToRetryAt time.Time,
	durationMs, numRetry int64,
	nodeTracesRaw []byte,
	requestBody []byte,
) (failure.Failure, error) {
	traces, err := decodeNodeTraces(nodeTracesRaw)
	if err != nil {
		return failure.Failure{}, err
	}
	return failure.Failure{
		RequestID:      requestID,
		NamespaceID:    namespaceID,
		GraphID:        graphID,
		Method:         method,
		Path:           path,
		StartedAt:      startedAt,
		FinishedAt:     finishedAt,
		DurationMs:     durationMs,
		Status:         failure.Status(status),
		Error:          errorMsg,
		NodeTraces:     traces,
		ReadyToRetryAt: readyToRetryAt,
		NumRetry:       numRetry,
		RequestBody:    requestBody,
	}, nil
}
