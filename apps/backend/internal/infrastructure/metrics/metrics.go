// Package metrics provides Prometheus metrics for business logic monitoring.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Session metrics
var (
	// SessionsCreated counts the total number of sessions created.
	SessionsCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "mindhit_sessions_created_total",
			Help: "Total number of sessions created",
		},
	)

	// SessionsCompleted counts the total number of sessions completed by status.
	SessionsCompleted = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mindhit_sessions_completed_total",
			Help: "Total number of sessions completed",
		},
		[]string{"status"}, // "success", "failed"
	)

	// SessionDuration observes the duration of sessions in seconds.
	SessionDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "mindhit_session_duration_seconds",
			Help:    "Session duration in seconds",
			Buckets: []float64{60, 300, 600, 1800, 3600, 7200},
		},
	)
)

// Event metrics
var (
	// EventsReceived counts the total number of events received by type.
	EventsReceived = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mindhit_events_received_total",
			Help: "Total number of events received by type",
		},
		[]string{"event_type"}, // "page_visit", "scroll", "highlight", "click", "page_leave"
	)

	// EventBatchSize observes the number of events per batch.
	EventBatchSize = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "mindhit_event_batch_size",
			Help:    "Number of events per batch",
			Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500},
		},
	)
)

// AI processing metrics
var (
	// AIRequestsTotal counts the total number of AI API requests.
	AIRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mindhit_ai_requests_total",
			Help: "Total number of AI API requests",
		},
		[]string{"provider", "operation", "status"}, // provider: openai/gemini/claude, operation: tag_extraction/mindmap_generation, status: success/error
	)

	// AIProcessingDuration observes the AI processing duration in seconds.
	AIProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mindhit_ai_processing_duration_seconds",
			Help:    "AI processing duration in seconds",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
		},
		[]string{"provider", "operation"},
	)

	// AITokensUsed counts the total number of AI tokens used.
	AITokensUsed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mindhit_ai_tokens_used_total",
			Help: "Total number of AI tokens used",
		},
		[]string{"provider", "token_type"}, // token_type: input/output
	)

	// AIProcessingErrors counts the total number of AI processing errors.
	AIProcessingErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mindhit_ai_processing_errors_total",
			Help: "Total number of AI processing errors",
		},
		[]string{"provider", "operation", "error_type"},
	)
)

// Worker/Job metrics
var (
	// WorkerJobsProcessed counts the total number of worker jobs processed.
	WorkerJobsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mindhit_worker_jobs_processed_total",
			Help: "Total number of worker jobs processed",
		},
		[]string{"job_type", "status"}, // job_type: session_processing/cleanup/tag_extraction/mindmap_generation, status: success/failed
	)

	// WorkerJobDuration observes the worker job processing duration in seconds.
	WorkerJobDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mindhit_worker_job_duration_seconds",
			Help:    "Worker job processing duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 5, 10, 30, 60, 300},
		},
		[]string{"job_type"},
	)

	// WorkerJobsInQueue tracks the number of jobs currently in queue.
	WorkerJobsInQueue = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mindhit_worker_jobs_in_queue",
			Help: "Number of jobs currently in queue",
		},
		[]string{"queue", "state"}, // state: pending/active/scheduled/retry
	)

	// WorkerJobRetries counts the total number of job retries.
	WorkerJobRetries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mindhit_worker_job_retries_total",
			Help: "Total number of job retries",
		},
		[]string{"job_type"},
	)
)

// Mindmap metrics
var (
	// MindmapsGenerated counts the total number of mindmaps generated.
	MindmapsGenerated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mindhit_mindmaps_generated_total",
			Help: "Total number of mindmaps generated",
		},
		[]string{"status"}, // "success", "failed"
	)

	// MindmapNodeCount observes the number of nodes per mindmap.
	MindmapNodeCount = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "mindhit_mindmap_node_count",
			Help:    "Number of nodes per mindmap",
			Buckets: []float64{5, 10, 25, 50, 100, 250, 500},
		},
	)

	// MindmapEdgeCount observes the number of edges per mindmap.
	MindmapEdgeCount = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "mindhit_mindmap_edge_count",
			Help:    "Number of edges per mindmap",
			Buckets: []float64{5, 10, 25, 50, 100, 250, 500},
		},
	)
)

// Database metrics
var (
	// DBQueryDuration observes the database query duration in seconds.
	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mindhit_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"operation"}, // "select", "insert", "update", "delete"
	)

	// DBConnectionsActive tracks the number of active database connections.
	DBConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "mindhit_db_connections_active",
			Help: "Number of active database connections",
		},
	)

	// DBConnectionsIdle tracks the number of idle database connections.
	DBConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "mindhit_db_connections_idle",
			Help: "Number of idle database connections",
		},
	)
)

// Redis/Cache metrics
var (
	// RedisCacheOperations counts the total number of Redis cache operations.
	RedisCacheOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mindhit_redis_cache_operations_total",
			Help: "Total number of Redis cache operations",
		},
		[]string{"operation", "result"}, // operation: get/set/delete, result: hit/miss/success/error
	)

	// RedisOperationDuration observes the Redis operation duration in seconds.
	RedisOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mindhit_redis_operation_duration_seconds",
			Help:    "Redis operation duration in seconds",
			Buckets: []float64{.0001, .0005, .001, .005, .01, .05, .1},
		},
		[]string{"operation"},
	)
)

// Auth metrics
var (
	// AuthAttempts counts the total number of authentication attempts.
	AuthAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mindhit_auth_attempts_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"method", "status"}, // method: password/google/refresh, status: success/failed
	)

	// AuthTokensIssued counts the total number of tokens issued.
	AuthTokensIssued = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mindhit_auth_tokens_issued_total",
			Help: "Total number of tokens issued",
		},
		[]string{"token_type"}, // "access", "refresh"
	)
)

// Subscription/Usage metrics
var (
	// SubscriptionsByPlan tracks the number of active subscriptions by plan.
	SubscriptionsByPlan = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mindhit_subscriptions_by_plan",
			Help: "Number of active subscriptions by plan",
		},
		[]string{"plan"}, // "free", "pro", "enterprise"
	)

	// TokenUsageDaily counts the daily token usage by user plan.
	TokenUsageDaily = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mindhit_token_usage_daily_total",
			Help: "Daily token usage by user plan",
		},
		[]string{"plan"},
	)

	// UsageLimitExceeded counts the number of times usage limit was exceeded.
	UsageLimitExceeded = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mindhit_usage_limit_exceeded_total",
			Help: "Number of times usage limit was exceeded",
		},
		[]string{"plan", "limit_type"}, // limit_type: daily/monthly
	)
)
