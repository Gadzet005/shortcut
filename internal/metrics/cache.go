package metrics

import (
	"github.com/Gadzet005/shortcut/internal/domain/graph"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type CacheMetrics struct {
	hits    *prometheus.CounterVec
	misses  *prometheus.CounterVec
	inserts *prometheus.CounterVec
}

func NewCacheMetrics(serviceName string) *CacheMetrics {
	constLabels := prometheus.Labels{"service": serviceName}
	labels := []string{"node_id"}
	return &CacheMetrics{
		hits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "shortcut_cache_hits_total",
				Help:        "Total number of node cache hits",
				ConstLabels: constLabels,
			},
			labels,
		),
		misses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "shortcut_cache_misses_total",
				Help:        "Total number of node cache misses",
				ConstLabels: constLabels,
			},
			labels,
		),
		inserts: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "shortcut_cache_inserts_total",
				Help:        "Total number of node cache inserts",
				ConstLabels: constLabels,
			},
			labels,
		),
	}
}

func (m *CacheMetrics) RecordHit(nodeID graph.NodeID) {
	m.hits.WithLabelValues(nodeID.String()).Inc()
}

func (m *CacheMetrics) RecordMiss(nodeID graph.NodeID) {
	m.misses.WithLabelValues(nodeID.String()).Inc()
}

func (m *CacheMetrics) RecordInsert(nodeID graph.NodeID) {
	m.inserts.WithLabelValues(nodeID.String()).Inc()
}
