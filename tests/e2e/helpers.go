package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type traceResponse struct {
	RequestID   string              `json:"request_id"`
	NamespaceID string              `json:"namespace_id"`
	GraphID     string              `json:"graph_id"`
	Method      string              `json:"method"`
	Path        string              `json:"path"`
	DurationMs  int64               `json:"duration_ms"`
	Status      string              `json:"status"`
	Error       string              `json:"error,omitempty"`
	NodeTraces  []nodeTraceResponse `json:"node_traces"`
}

type nodeTraceResponse struct {
	NodeID     string `json:"node_id"`
	DurationMs int64  `json:"duration_ms"`
	StatusCode int    `json:"status_code,omitempty"`
	RetryCount int    `json:"retry_count,omitempty"`
	Error      string `json:"error,omitempty"`
}

func getTrace(t *testing.T, shortcutURL string, requestID string) traceResponse {
	t.Helper()
	require.NotEmpty(t, requestID, "X-Request-Id header must be set")
	resp, err := http.Get(shortcutURL + "/trace/" + requestID)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "trace should exist for request_id=%s", requestID)
	var result traceResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	return result
}

// findNodeTrace returns the NodeTrace with the given nodeID, or fails the test if not found.
func findNodeTrace(t *testing.T, tr traceResponse, nodeID string) nodeTraceResponse {
	t.Helper()
	for _, nt := range tr.NodeTraces {
		if nt.NodeID == nodeID {
			return nt
		}
	}
	t.Fatalf("node trace not found for nodeID=%q (got %d node traces)", nodeID, len(tr.NodeTraces))
	return nodeTraceResponse{}
}
