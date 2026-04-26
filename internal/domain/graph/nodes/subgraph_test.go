package graphnodes

import (
	"errors"
	"testing"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	mockgraph "github.com/Gadzet005/shortcut/internal/domain/graph/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSubgraphNodeExecutor_Success(t *testing.T) {
	g := mockgraph.NewGraph(t)
	outItems := map[graph.ItemID]graph.Item{"result": {Data: []byte("ok")}}
	g.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(outItems, nil)

	exec := NewSubgraphNodeExecutor(g)
	resp, err := exec.Run(t.Context(), zap.NewNop(), graph.NodeExecutorRequest{
		Items: map[graph.ItemID]graph.Item{"in": {Data: []byte("x")}},
	})
	require.NoError(t, err)
	require.Equal(t, outItems, resp.Items)
}

func TestSubgraphNodeExecutor_GraphError(t *testing.T) {
	g := mockgraph.NewGraph(t)
	sentinel := errors.New("graph exploded")
	g.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, sentinel)

	exec := NewSubgraphNodeExecutor(g)
	_, err := exec.Run(t.Context(), zap.NewNop(), graph.NodeExecutorRequest{})
	require.ErrorContains(t, err, sentinel.Error())
}
