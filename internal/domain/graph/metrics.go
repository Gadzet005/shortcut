package graph

import "time"

type GraphMetrics interface {
	ObserveRun(namespaceID NamespaceID, graphID ID, duration time.Duration, err error)
}
