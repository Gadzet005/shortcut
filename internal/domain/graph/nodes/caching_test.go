package graphnodes

import (
	"errors"
	"testing"
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	mockgraph "github.com/Gadzet005/shortcut/internal/domain/graph/mocks"
	mocknodes "github.com/Gadzet005/shortcut/internal/domain/graph/nodes/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCachingExecutor_CacheHit(t *testing.T) {
	repo := mocknodes.NewCacheRepo(t)
	inner := mockgraph.NewNodeExecutor(t)

	cached := graph.NodeExecutorResponse{
		Items: map[graph.ItemID]graph.Item{"out": {Data: []byte("cached-data")}},
	}
	repo.EXPECT().Get(mock.Anything, mock.Anything).Return(cached, true, nil)

	exec := newCachingExecutor(inner, repo)
	resp, err := exec.Run(t.Context(), zap.NewNop(), graph.NodeExecutorRequest{
		Items: map[graph.ItemID]graph.Item{"in": {Data: []byte("x")}},
	})
	require.NoError(t, err)
	require.Equal(t, []byte("cached-data"), resp.Items["out"].Data)
	require.Equal(t, true, resp.Meta["cached"])
}

func TestCachingExecutor_CacheMiss_StoresResult(t *testing.T) {
	repo := mocknodes.NewCacheRepo(t)
	inner := mockgraph.NewNodeExecutor(t)

	fresh := graph.NodeExecutorResponse{
		Items: map[graph.ItemID]graph.Item{"out": {Data: []byte("fresh")}},
	}

	repo.EXPECT().Get(mock.Anything, mock.Anything).Once().Return(graph.NodeExecutorResponse{}, false, nil)
	repo.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
	repo.EXPECT().Get(mock.Anything, mock.Anything).Once().Return(fresh, true, nil)
	inner.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).Once().Return(fresh, nil)

	exec := newCachingExecutor(inner, repo)
	req := graph.NodeExecutorRequest{Items: map[graph.ItemID]graph.Item{"in": {Data: []byte("x")}}}

	resp, err := exec.Run(t.Context(), zap.NewNop(), req)
	require.NoError(t, err)
	require.Equal(t, []byte("fresh"), resp.Items["out"].Data)
	require.NotEqual(t, true, resp.Meta["cached"])

	resp2, err := exec.Run(t.Context(), zap.NewNop(), req)
	require.NoError(t, err)
	require.Equal(t, true, resp2.Meta["cached"])
}

func TestCachingExecutor_EndpointOverride_BypassesCache(t *testing.T) {
	repo := mocknodes.NewCacheRepo(t)
	inner := mockgraph.NewNodeExecutor(t)

	inner.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).Return(graph.NodeExecutorResponse{}, nil)

	override := "localhost:9999"
	exec := newCachingExecutor(inner, repo)
	_, err := exec.Run(t.Context(), zap.NewNop(), graph.NodeExecutorRequest{
		Items:            map[graph.ItemID]graph.Item{"in": {Data: []byte("x")}},
		EndpointOverride: &override,
	})
	require.NoError(t, err)
}

func TestCachingExecutor_CacheGetError_FallsBackToInner(t *testing.T) {
	repo := mocknodes.NewCacheRepo(t)
	inner := mockgraph.NewNodeExecutor(t)

	repo.EXPECT().Get(mock.Anything, mock.Anything).Return(graph.NodeExecutorResponse{}, false, errors.New("redis unavailable"))
	repo.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	inner.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).Return(graph.NodeExecutorResponse{}, nil)

	exec := newCachingExecutor(inner, repo)
	_, err := exec.Run(t.Context(), zap.NewNop(), graph.NodeExecutorRequest{})
	require.NoError(t, err)
}

func TestCachingExecutor_CacheSetError_ReturnsResponseAnyway(t *testing.T) {
	repo := mocknodes.NewCacheRepo(t)
	inner := mockgraph.NewNodeExecutor(t)

	fresh := graph.NodeExecutorResponse{
		Items: map[graph.ItemID]graph.Item{"out": {Data: []byte("ok")}},
	}
	repo.EXPECT().Get(mock.Anything, mock.Anything).Return(graph.NodeExecutorResponse{}, false, nil)
	repo.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("write failed"))
	inner.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).Return(fresh, nil)

	exec := newCachingExecutor(inner, repo)
	resp, err := exec.Run(t.Context(), zap.NewNop(), graph.NodeExecutorRequest{})
	require.NoError(t, err)
	require.Equal(t, []byte("ok"), resp.Items["out"].Data)
}

func TestCachingExecutor_InnerError_NotCached(t *testing.T) {
	repo := mocknodes.NewCacheRepo(t)
	inner := mockgraph.NewNodeExecutor(t)

	sentinel := errors.New("inner failure")
	repo.EXPECT().Get(mock.Anything, mock.Anything).Return(graph.NodeExecutorResponse{}, false, nil)
	inner.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).Return(graph.NodeExecutorResponse{}, sentinel)

	exec := newCachingExecutor(inner, repo)
	_, err := exec.Run(t.Context(), zap.NewNop(), graph.NodeExecutorRequest{})
	require.ErrorIs(t, err, sentinel)
}

func TestHashItems_Deterministic(t *testing.T) {
	items1 := map[graph.ItemID]graph.Item{
		"b": {Data: []byte("beta")},
		"a": {Data: []byte("alpha")},
	}
	items2 := map[graph.ItemID]graph.Item{
		"a": {Data: []byte("alpha")},
		"b": {Data: []byte("beta")},
	}
	require.Equal(t, hashItems(items1), hashItems(items2))
}

func TestHashItems_DifferentDataProducesDifferentHash(t *testing.T) {
	a := map[graph.ItemID]graph.Item{"k": {Data: []byte("v1")}}
	b := map[graph.ItemID]graph.Item{"k": {Data: []byte("v2")}}
	require.NotEqual(t, hashItems(a), hashItems(b))
}

func newCachingExecutor(inner graph.NodeExecutor, repo CacheRepo) *CachingExecutor {
	return NewCachingExecutor(inner, "node1", "graph-hash", time.Minute, repo)
}
