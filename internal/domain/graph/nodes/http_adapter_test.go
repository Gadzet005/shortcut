package graphnodes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	shortcutapi "github.com/Gadzet005/shortcut/pkg/shortcut/api"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestHTTPAdapterNodeExecutor_Success(t *testing.T) {
	responseBody := []byte(`{"key":"value"}`)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom", "yes")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(responseBody)
	}))
	defer srv.Close()

	exec := NewHTTPAdapterNodeExecutor(resty.New(), Endpoint{URL: srv.URL + "/path"})
	req := graph.NodeExecutorRequest{
		Items: map[graph.ItemID]graph.Item{
			httpAdapterRequestItemID: encodeHTTPRequest(t, shortcutapi.HttpRequest{
				Method: http.MethodPost,
				Headers: map[string][]string{
					"Accept": {"application/json"},
				},
			}),
		},
	}

	resp, err := exec.Run(t.Context(), zap.NewNop(), req)
	require.NoError(t, err)

	rawItem, ok := resp.Items[httpAdapterResponseItemID]
	require.True(t, ok)

	var httpResp shortcutapi.HttpResponse
	require.NoError(t, json.Unmarshal(rawItem.Data, &httpResp))
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.Equal(t, responseBody, httpResp.Body)
	require.Equal(t, http.StatusOK, resp.Meta["status_code"])
}

func TestHTTPAdapterNodeExecutor_MissingRequestItem(t *testing.T) {
	exec := NewHTTPAdapterNodeExecutor(resty.New(), Endpoint{URL: "http://localhost"})
	_, err := exec.Run(t.Context(), zap.NewNop(), graph.NodeExecutorRequest{
		Items: map[graph.ItemID]graph.Item{},
	})
	require.ErrorContains(t, err, "http_request item not found")
}

func TestHTTPAdapterNodeExecutor_InvalidRequestJSON(t *testing.T) {
	exec := NewHTTPAdapterNodeExecutor(resty.New(), Endpoint{URL: "http://localhost"})
	_, err := exec.Run(t.Context(), zap.NewNop(), graph.NodeExecutorRequest{
		Items: map[graph.ItemID]graph.Item{
			httpAdapterRequestItemID: {Data: []byte("not-json")},
		},
	})
	require.ErrorContains(t, err, "unmarshal http_request")
}

func TestHTTPAdapterNodeExecutor_4xxError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad input"}`))
	}))
	defer srv.Close()

	exec := NewHTTPAdapterNodeExecutor(resty.New(), Endpoint{URL: srv.URL})
	req := graph.NodeExecutorRequest{
		Items: map[graph.ItemID]graph.Item{
			httpAdapterRequestItemID: encodeHTTPRequest(t, shortcutapi.HttpRequest{Method: http.MethodGet}),
		},
	}

	_, err := exec.Run(t.Context(), zap.NewNop(), req)
	require.Error(t, err)

	var nodeErr *graph.NodeError
	require.ErrorAs(t, err, &nodeErr)
	require.Equal(t, graph.ErrCodeBadRequest, nodeErr.Code)
}

func TestHTTPAdapterNodeExecutor_5xxError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"oops"}`))
	}))
	defer srv.Close()

	exec := NewHTTPAdapterNodeExecutor(resty.New(), Endpoint{URL: srv.URL, RetriesNum: 0})
	req := graph.NodeExecutorRequest{
		Items: map[graph.ItemID]graph.Item{
			httpAdapterRequestItemID: encodeHTTPRequest(t, shortcutapi.HttpRequest{Method: http.MethodGet}),
		},
	}

	_, err := exec.Run(t.Context(), zap.NewNop(), req)
	require.Error(t, err)

	var nodeErr *graph.NodeError
	require.ErrorAs(t, err, &nodeErr)
	require.Equal(t, graph.ErrCodeInternal, nodeErr.Code)
}

func encodeHTTPRequest(t *testing.T, req shortcutapi.HttpRequest) graph.Item {
	t.Helper()
	b, err := json.Marshal(req)
	require.NoError(t, err)
	return graph.Item{Data: b}
}
