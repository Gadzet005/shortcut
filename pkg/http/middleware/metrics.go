package httpmiddleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	defaultEndpointName = "unknown"
	defaultNamespace    = "unknown"
)

func Metrics(serviceName string) gin.HandlerFunc {
	return metricsHandler(newHTTPMetrics(serviceName, promauto.With(prometheus.DefaultRegisterer)))
}

func metricsHandler(m *httpServiceMetrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				m.panicsTotal.Inc()
				panic(err)
			}
		}()

		startTime := time.Now()

		namespace := c.Param("namespace_id")
		if namespace == "" {
			namespace = defaultNamespace
		}

		path := c.Param("path")
		if path == "" {
			path = "/"
		}

		c.Next()

		duration := time.Since(startTime).Seconds()
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = defaultEndpointName
		}

		method := c.Request.Method
		m.requestsCnt.WithLabelValues(method, endpoint, path, namespace).Inc()
		m.requestQuantiles.WithLabelValues(method, endpoint, path, namespace).Observe(duration)
		m.codesTotal.WithLabelValues(method, endpoint, strconv.Itoa(c.Writer.Status()), path, namespace).Inc()
	}
}

func newHTTPMetrics(serviceName string, factory promauto.Factory) *httpServiceMetrics {
	constLabels := prometheus.Labels{
		"service": serviceName,
	}

	return &httpServiceMetrics{
		requestsCnt: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "http_requests_total",
				Help:        "Total number of HTTP requests",
				ConstLabels: constLabels,
			},
			[]string{"method", "endpoint", "path", "namespace"},
		),
		requestQuantiles: factory.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:        "http_request_duration_quantiles_seconds",
				Help:        "Quantiles of HTTP request duration",
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
			[]string{"method", "endpoint", "path", "namespace"},
		),
		requestSize: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "http_request_size_bytes",
				Help:        "Size of HTTP requests in bytes",
				ConstLabels: constLabels,
				Buckets:     []float64{100, 1000, 10000, 100000, 1000000, 10000000},
			},
			[]string{"method", "endpoint", "path", "namespace"},
		),
		responseSize: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "http_response_size_bytes",
				Help:        "Size of HTTP responses in bytes",
				ConstLabels: constLabels,
				Buckets:     []float64{100, 1000, 10000, 100000, 1000000, 10000000},
			},
			[]string{"method", "endpoint", "path", "namespace"},
		),
		codesTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "http_codes_total",
				Help:        "Total number of HTTP errors by code",
				ConstLabels: constLabels,
			},
			[]string{"method", "endpoint", "code", "path", "namespace"},
		),
		panicsTotal: factory.NewCounter(
			prometheus.CounterOpts{
				Name:        "http_panics_total",
				Help:        "Total number of HTTP panics",
				ConstLabels: constLabels,
			},
		),
	}
}

type httpServiceMetrics struct {
	requestsCnt      *prometheus.CounterVec
	requestQuantiles *prometheus.SummaryVec
	responseSize     *prometheus.HistogramVec
	requestSize      *prometheus.HistogramVec
	codesTotal       *prometheus.CounterVec
	panicsTotal      prometheus.Counter
}
