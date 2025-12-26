package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mindhit_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mindhit_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "mindhit_http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
	)

	sessionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "mindhit_sessions_active",
			Help: "Number of active recording sessions",
		},
	)

	eventsProcessed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "mindhit_events_processed_total",
			Help: "Total number of events processed",
		},
	)
)

// Metrics returns a Gin middleware that collects HTTP metrics.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}

// IncrementEventsProcessed increments the events counter.
func IncrementEventsProcessed(count int) {
	eventsProcessed.Add(float64(count))
}

// SetActiveSessions sets the active sessions gauge.
func SetActiveSessions(count int) {
	sessionsActive.Set(float64(count))
}
