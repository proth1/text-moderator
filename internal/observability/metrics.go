package observability

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Control: OBS-001 (Observability - Prometheus metrics)

// Metrics holds all Prometheus metrics for the service.
type Metrics struct {
	// HTTP request metrics
	RequestDuration *prometheus.HistogramVec
	RequestTotal    *prometheus.CounterVec

	// Moderation-specific metrics
	ModerationTotal    *prometheus.CounterVec
	ModerationDuration *prometheus.HistogramVec
	ModerationActions  *prometheus.CounterVec
	ClassificationCacheHits   prometheus.Counter
	ClassificationCacheMisses prometheus.Counter

	// Provider metrics
	ProviderRequestTotal    *prometheus.CounterVec
	ProviderRequestDuration *prometheus.HistogramVec
	ProviderFailures        *prometheus.CounterVec

	// Policy evaluation metrics
	PolicyEvaluationTotal    *prometheus.CounterVec
	PolicyEvaluationDuration prometheus.Histogram

	// Webhook metrics
	WebhookDeliveryTotal    *prometheus.CounterVec
	WebhookDeliveryDuration prometheus.Histogram

	// Review metrics
	ReviewQueueSize prometheus.Gauge
	ReviewTotal     *prometheus.CounterVec
}

// NewMetrics creates and registers all Prometheus metrics.
func NewMetrics(serviceName string) *Metrics {
	return &Metrics{
		RequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
			ConstLabels: prometheus.Labels{"service": serviceName},
		}, []string{"method", "path", "status"}),

		RequestTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name:        "http_requests_total",
			Help:        "Total number of HTTP requests",
			ConstLabels: prometheus.Labels{"service": serviceName},
		}, []string{"method", "path", "status"}),

		ModerationTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "moderation_requests_total",
			Help: "Total number of moderation requests",
		}, []string{"action", "provider"}),

		ModerationDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "moderation_duration_seconds",
			Help:    "End-to-end moderation pipeline duration",
			Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		}, []string{"provider", "cache_hit"}),

		ModerationActions: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "moderation_actions_total",
			Help: "Total moderation actions by type",
		}, []string{"action"}),

		ClassificationCacheHits: promauto.NewCounter(prometheus.CounterOpts{
			Name: "classification_cache_hits_total",
			Help: "Total classification cache hits",
		}),

		ClassificationCacheMisses: promauto.NewCounter(prometheus.CounterOpts{
			Name: "classification_cache_misses_total",
			Help: "Total classification cache misses",
		}),

		ProviderRequestTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "provider_requests_total",
			Help: "Total provider classification requests",
		}, []string{"provider", "status"}),

		ProviderRequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "provider_request_duration_seconds",
			Help:    "Provider API request duration",
			Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30},
		}, []string{"provider"}),

		ProviderFailures: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "provider_failures_total",
			Help: "Total provider failures triggering fallback",
		}, []string{"provider"}),

		PolicyEvaluationTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "policy_evaluation_total",
			Help: "Total policy evaluations",
		}, []string{"policy_name", "action"}),

		PolicyEvaluationDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "policy_evaluation_duration_seconds",
			Help:    "Policy evaluation duration",
			Buckets: prometheus.DefBuckets,
		}),

		WebhookDeliveryTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "webhook_deliveries_total",
			Help: "Total webhook delivery attempts",
		}, []string{"event_type", "status"}),

		WebhookDeliveryDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "webhook_delivery_duration_seconds",
			Help:    "Webhook delivery duration",
			Buckets: []float64{0.1, 0.5, 1, 5, 10, 30},
		}),

		ReviewQueueSize: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "review_queue_size",
			Help: "Current number of items in the review queue",
		}),

		ReviewTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "review_actions_total",
			Help: "Total review actions by type",
		}, []string{"action"}),
	}
}

// MetricsMiddleware returns a Gin middleware that records HTTP request metrics.
func MetricsMiddleware(m *Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		m.RequestDuration.WithLabelValues(c.Request.Method, path, status).Observe(duration)
		m.RequestTotal.WithLabelValues(c.Request.Method, path, status).Inc()
	}
}

// PrometheusHandler returns a Gin handler that exposes Prometheus metrics.
func PrometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
