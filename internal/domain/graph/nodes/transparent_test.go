package graphnodes

import (
	"testing"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestTransparentNodeExecutor(t *testing.T) {
	exec := NewTransparentNodeExecutor()
	items := map[graph.ItemID]graph.Item{
		"a": {Data: []byte("hello")},
		"b": {Data: []byte("world")},
	}

	resp, err := exec.Run(t.Context(), zap.NewNop(), graph.NodeExecutorRequest{Items: items})
	require.NoError(t, err)
	require.Equal(t, items, resp.Items)
}

func TestTransparentNodeExecutor_EmptyItems(t *testing.T) {
	exec := NewTransparentNodeExecutor()
	resp, err := exec.Run(t.Context(), zap.NewNop(), graph.NodeExecutorRequest{Items: map[graph.ItemID]graph.Item{}})
	require.NoError(t, err)
	require.Empty(t, resp.Items)
}
