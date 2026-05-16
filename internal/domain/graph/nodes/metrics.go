package graphnodes

import (
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
)

type NodeMetrics interface {
	ObserveRun(namespaceID graph.NamespaceID, graphID graph.ID, nodeID graph.NodeID, nodeType string, duration time.Duration, err error)
}

type CacheMetrics interface {
	RecordHit(nodeID graph.NodeID)
	RecordMiss(nodeID graph.NodeID)
	RecordInsert(nodeID graph.NodeID)
}
