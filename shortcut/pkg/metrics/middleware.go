package metrics

import (
    "strconv"
    "time"
     
	"github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const defaultEndpointName = "unknown"

type HTTPServiceMetrics struct {
    requestsCnt  *prometheus.CounterVec
	requestQuantiles *prometheus.SummaryVec
    responseSize *prometheus.HistogramVec
	requestSize *prometheus.HistogramVec
    codesTotal *prometheus.CounterVec
    panicsTotal prometheus.Counter
}

func NewHTTPServiceMetrics(serviceName string) *HTTPServiceMetrics {
    constLabels := prometheus.Labels{
        "service": serviceName,
    }
    
    return &HTTPServiceMetrics{
        requestsCnt: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name:        "http_requests_total",
                Help:        "Total number of HTTP requests",
                ConstLabels: constLabels,
            },
            []string{"method", "endpoint"},
        ),
        requestQuantiles: promauto.NewSummaryVec(
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
                MaxAge:      time.Minute,
                AgeBuckets:  5,
                BufCap:      500,
            },
            []string{"method", "endpoint"},
        ),
		requestSize: promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:        "http_request_size_bytes",
                Help:        "Size of HTTP requests in bytes",
                ConstLabels: constLabels,
                Buckets:     []float64{100, 1000, 10000, 100000, 1000000, 10000000},
            },
            []string{"method", "endpoint"},
        ),
        responseSize: promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:        "http_response_size_bytes",
                Help:        "Size of HTTP responses in bytes",
                ConstLabels: constLabels,
                Buckets:     []float64{100, 1000, 10000, 100000, 1000000, 10000000},
            },
            []string{"method", "endpoint"},
        ),
        codesTotal: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name:        "http_codes_total",
                Help:        "Total number of HTTP errors by code",
                ConstLabels: constLabels,
            },
            []string{"method", "endpoint", "code"},
        ),
        panicsTotal: promauto.NewCounter(
            prometheus.CounterOpts{
                Name:        "http_panics_total",
                Help:        "Total number of HTTP panics",
                ConstLabels: constLabels,
            },
        ),
    }
}

func (m *HTTPServiceMetrics) MetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
		defer func() {
            if err := recover(); err != nil {
                m.panicsTotal.Inc()
                panic(err)
            }
        }()

        startTime := time.Now()

        c.Next()

        duration := time.Since(startTime).Seconds()
        endpoint := c.FullPath()
		if endpoint == "" {
            endpoint = defaultEndpointName
        }

		m.requestsCnt.WithLabelValues(c.Request.Method, endpoint).Inc()
		m.requestQuantiles.WithLabelValues(c.Request.Method, endpoint).Observe(duration)

		if c.Request != nil {
			m.requestSize.WithLabelValues(c.Request.Method, endpoint).Observe(float64(c.Request.ContentLength))
			if c.Request.Response != nil {
				m.codesTotal.WithLabelValues(c.Request.Method, endpoint, strconv.Itoa(c.Request.Response.StatusCode)).Inc() // TODO: fix getting sizes and status code
				m.responseSize.WithLabelValues(c.Request.Method, endpoint).Observe(float64(c.Request.Response.ContentLength))
			}
		}
    }
}
