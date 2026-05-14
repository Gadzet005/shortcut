package metrics

import (
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/graph"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type GraphMetrics struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.SummaryVec
	errorsTotal     *prometheus.CounterVec
}

func NewGraphMetrics(serviceName string) *GraphMetrics {
	constLabels := prometheus.Labels{"service": serviceName}
	labels := []string{"namespace_id", "graph_id"}
	return &GraphMetrics{
		requestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "shortcut_graph_requests_total",
				Help:        "Total number of graph executions",
				ConstLabels: constLabels,
			},
			labels,
		),
		requestDuration: promauto.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:        "shortcut_graph_duration_seconds",
				Help:        "Quantiles of graph execution duration",
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
				Name:        "shortcut_graph_errors_total",
				Help:        "Total number of graph executions that ended with an error",
				ConstLabels: constLabels,
			},
			labels,
		),
	}
}

func (m *GraphMetrics) ObserveRun(namespaceID graph.NamespaceID, graphID graph.ID, duration time.Duration, err error) {
	ns := namespaceID.String()
	gID := graphID.String()
	m.requestsTotal.WithLabelValues(ns, gID).Inc()
	m.requestDuration.WithLabelValues(ns, gID).Observe(duration.Seconds())
	if err != nil {
		m.errorsTotal.WithLabelValues(ns, gID).Inc()
	}
}
