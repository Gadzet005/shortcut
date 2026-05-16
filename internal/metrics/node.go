package metrics

import (
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type NodeMetrics struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.SummaryVec
	errorsTotal     *prometheus.CounterVec
}

func NewNodeMetrics(serviceName string) *NodeMetrics {
	constLabels := prometheus.Labels{"service": serviceName}
	labels := []string{"namespace", "graph_id", "node_id", "node_type"}
	return &NodeMetrics{
		requestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "shortcut_node_requests_total",
				Help:        "Total number of node executions",
				ConstLabels: constLabels,
			},
			labels,
		),
		requestDuration: promauto.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:        "shortcut_node_duration_seconds",
				Help:        "Quantiles of node execution duration",
				ConstLabels: constLabels,
				Objectives: map[float64]float64{
					0.5:  0.05,
					0.9:  0.01,
					0.95: 0.005,
					0.99: 0.001,
				},
				MaxAge:     time.Minute,
				AgeBuckets: 5,
				BufCap:     500,
			},
			labels,
		),
		errorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "shortcut_node_errors_total",
				Help:        "Total number of node executions that ended with an error",
				ConstLabels: constLabels,
			},
			labels,
		),
	}
}

func (m *NodeMetrics) ObserveRun(namespaceID graph.NamespaceID, graphID graph.ID, nodeID graph.NodeID, nodeType string, duration time.Duration, err error) {
	m.requestsTotal.WithLabelValues(namespaceID.String(), graphID.String(), nodeID.String(), nodeType).Inc()
	m.requestDuration.WithLabelValues(namespaceID.String(), graphID.String(), nodeID.String(), nodeType).Observe(duration.Seconds())
	if err != nil {
		m.errorsTotal.WithLabelValues(namespaceID.String(), graphID.String(), nodeID.String(), nodeType).Inc()
	}
}
