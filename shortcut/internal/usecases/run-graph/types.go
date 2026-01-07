package rungraph

import (
	"context"

	"github.com/Gadzet005/shortcut/shortcut/internal/domain/graph"
)

type RunGraphRequest struct {
	GraphID graph.GraphID
	Data    []byte
}

type RunGraphResponse struct {
	Data []byte
}

type UseCase interface {
	RunGraph(ctx context.Context, input RunGraphRequest) (RunGraphResponse, error)
}

type nodeResult map[graph.ItemID]graph.Item
type nodeResults map[graph.NodeID]nodeResult

func (r nodeResults) Get(nodeID graph.NodeID, itemID graph.ItemID) (graph.Item, bool) {
	nodeResult, ok := r[nodeID]
	if !ok {
		return graph.Item{}, false
	}
	item, ok := nodeResult[itemID]
	return item, ok
}
