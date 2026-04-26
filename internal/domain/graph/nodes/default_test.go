package graphnodes

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	multipartutils "github.com/Gadzet005/shortcut/pkg/utils/multipart"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDefaultNodeExecutor_Success(t *testing.T) {
	outData := map[string][]byte{
		"result": []byte("processed"),
	}
	srv := multipartServer(t, outData)
	defer srv.Close()

	exec := NewDefaultNodeExecutor(resty.New(), Endpoint{URL: srv.URL})
	req := graph.NodeExecutorRequest{
		Items: map[graph.ItemID]graph.Item{
			"input": {Data: []byte("raw")},
		},
	}

	resp, err := exec.Run(t.Context(), zap.NewNop(), req)
	require.NoError(t, err)
	require.Equal(t, []byte("processed"), resp.Items["result"].Data)
	require.Equal(t, http.StatusOK, resp.Meta["status_code"])
}

func TestDefaultNodeExecutor_NonOKStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		wantCode graph.ErrorCode
	}{
		{"bad request", http.StatusBadRequest, graph.ErrCodeBadRequest},
		{"unauthorized", http.StatusUnauthorized, graph.ErrCodeUnauthorized},
		{"forbidden", http.StatusForbidden, graph.ErrCodeForbidden},
		{"not found", http.StatusNotFound, graph.ErrCodeNotFound},
		{"internal server error", http.StatusInternalServerError, graph.ErrCodeInternal},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(`{"error":"something"}`))
			}))
			defer srv.Close()

			exec := NewDefaultNodeExecutor(resty.New(), Endpoint{URL: srv.URL, RetriesNum: 0})
			_, err := exec.Run(t.Context(), zap.NewNop(), graph.NodeExecutorRequest{})
			require.Error(t, err)

			var nodeErr *graph.NodeError
			require.ErrorAs(t, err, &nodeErr)
			require.Equal(t, tc.wantCode, nodeErr.Code)
		})
	}
}

func TestDefaultNodeExecutor_EndpointOverride(t *testing.T) {
	outData := map[string][]byte{"x": []byte("overridden")}
	srv := multipartServer(t, outData)
	defer srv.Close()

	host := srv.Listener.Addr().String()
	exec := NewDefaultNodeExecutor(resty.New(), Endpoint{URL: "http://original-host/path"})
	resp, err := exec.Run(t.Context(), zap.NewNop(), graph.NodeExecutorRequest{
		Items:            map[graph.ItemID]graph.Item{},
		EndpointOverride: &host,
	})
	require.NoError(t, err)
	require.Equal(t, []byte("overridden"), resp.Items["x"].Data)
}

func multipartServer(t *testing.T, data map[string][]byte) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		var buf bytes.Buffer
		contentType, err := multipartutils.WriteMultipartData(&buf, data)
		require.NoError(t, err)
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buf.Bytes())
	}))
}
