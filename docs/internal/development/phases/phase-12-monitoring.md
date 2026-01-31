# Phase 12: 프로덕션 모니터링 시스템

## 개요

| 항목 | 내용 |
|-----|------|
| **목표** | 프로덕션 레벨 관측 가능성 (Observability) 구축 |
| **선행 조건** | Phase 11 (대시보드) 완료 |
| **예상 소요** | 4 Steps |
| **결과물** | Grafana 대시보드, 중앙 집중 로깅, 알림/온콜 시스템 |

> **Note**: Phase 5에서 기본 Prometheus 메트릭과 구조화된 로깅을 설정했습니다.
> 이 Phase에서는 프로덕션 환경에서의 모니터링, 대시보드 시각화, 알림 시스템을 구축합니다.

---

## 현재 인프라 상태 (Phase 5에서 구현됨)

### 이미 구현된 항목

| 항목 | 파일 위치 | 상태 |
|------|----------|------|
| Prometheus 메트릭 미들웨어 | `internal/infrastructure/middleware/metrics.go` | ✅ |
| 구조화된 로깅 | `internal/infrastructure/logger/logger.go` | ✅ |
| HTTP 로깅 미들웨어 | `internal/infrastructure/middleware/logging.go` | ✅ |
| Request ID 미들웨어 | `internal/infrastructure/middleware/request_id.go` | ✅ |
| Prometheus 서비스 | `infra/docker/docker-compose.yml` | ✅ |
| Grafana 서비스 | `infra/docker/docker-compose.yml` | ✅ |
| 기본 prometheus.yml | `infra/docker/prometheus.yml` | ✅ |

### 현재 포트 할당

| 서비스 | 포트 | 비고 |
|-------|------|------|
| API Server | 9000 | `host.docker.internal:9000` |
| Prometheus | 9091 | 9090은 asynqmon이 사용 |
| Grafana | 3010 | |
| Asynqmon | 9090 | Asynq 모니터링 UI |
| Alertmanager | 9093 | 알림 관리 |
| Loki | 3100 | 로그 수집 |
| PostgreSQL | 5433 | |
| Redis | 6380 | |

### 현재 메트릭 (middleware/metrics.go)

```go
// 이미 구현된 메트릭
mindhit_http_requests_total{method, path, status}
mindhit_http_request_duration_seconds{method, path}
mindhit_http_requests_in_flight
mindhit_sessions_active
mindhit_events_processed_total
```

---

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 12.1 | 비즈니스/Worker 메트릭 확장 | ✅ |
| 12.2 | Grafana 대시보드 프로비저닝 | ✅ |
| 12.3 | 로그 수집 시스템 (Loki) | ✅ |
| 12.4 | 알림 시스템 구성 | ✅ |

---

## Step 12.1: 비즈니스/Worker 메트릭 확장

### 목표

기존 HTTP 메트릭에 비즈니스 로직, AI 처리, Worker 관련 메트릭 추가

### 체크리스트

- [x] **비즈니스 메트릭 파일 생성**
  - [x] `internal/infrastructure/metrics/metrics.go`

    ```go
    // Package metrics provides Prometheus metrics for business logic monitoring.
    package metrics

    import (
        "github.com/prometheus/client_golang/prometheus"
        "github.com/prometheus/client_golang/prometheus/promauto"
    )

    // Session metrics
    var (
        SessionsCreated = promauto.NewCounter(
            prometheus.CounterOpts{
                Name: "mindhit_sessions_created_total",
                Help: "Total number of sessions created",
            },
        )

        SessionsCompleted = promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "mindhit_sessions_completed_total",
                Help: "Total number of sessions completed",
            },
            []string{"status"}, // "success", "failed"
        )

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
        EventsReceived = promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "mindhit_events_received_total",
                Help: "Total number of events received by type",
            },
            []string{"event_type"}, // "page_visit", "scroll", "highlight", "click"
        )

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
        AIRequestsTotal = promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "mindhit_ai_requests_total",
                Help: "Total number of AI API requests",
            },
            []string{"provider", "operation", "status"}, // provider: openai/gemini/claude, operation: tag_extraction/mindmap_generation, status: success/error
        )

        AIProcessingDuration = promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "mindhit_ai_processing_duration_seconds",
                Help:    "AI processing duration in seconds",
                Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
            },
            []string{"provider", "operation"},
        )

        AITokensUsed = promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "mindhit_ai_tokens_used_total",
                Help: "Total number of AI tokens used",
            },
            []string{"provider", "token_type"}, // token_type: input/output
        )

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
        WorkerJobsProcessed = promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "mindhit_worker_jobs_processed_total",
                Help: "Total number of worker jobs processed",
            },
            []string{"job_type", "status"}, // job_type: session_processing/cleanup/tag_extraction/mindmap_generation, status: success/failed/retried
        )

        WorkerJobDuration = promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "mindhit_worker_job_duration_seconds",
                Help:    "Worker job processing duration in seconds",
                Buckets: []float64{0.1, 0.5, 1, 5, 10, 30, 60, 300},
            },
            []string{"job_type"},
        )

        WorkerJobsInQueue = promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "mindhit_worker_jobs_in_queue",
                Help: "Number of jobs currently in queue",
            },
            []string{"queue", "state"}, // state: pending/active/scheduled/retry
        )

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
        MindmapsGenerated = promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "mindhit_mindmaps_generated_total",
                Help: "Total number of mindmaps generated",
            },
            []string{"status"}, // "success", "failed"
        )

        MindmapNodeCount = promauto.NewHistogram(
            prometheus.HistogramOpts{
                Name:    "mindhit_mindmap_node_count",
                Help:    "Number of nodes per mindmap",
                Buckets: []float64{5, 10, 25, 50, 100, 250, 500},
            },
        )

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
        DBQueryDuration = promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "mindhit_db_query_duration_seconds",
                Help:    "Database query duration in seconds",
                Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
            },
            []string{"operation"}, // "select", "insert", "update", "delete"
        )

        DBConnectionsActive = promauto.NewGauge(
            prometheus.GaugeOpts{
                Name: "mindhit_db_connections_active",
                Help: "Number of active database connections",
            },
        )

        DBConnectionsIdle = promauto.NewGauge(
            prometheus.GaugeOpts{
                Name: "mindhit_db_connections_idle",
                Help: "Number of idle database connections",
            },
        )
    )

    // Redis/Cache metrics
    var (
        RedisCacheOperations = promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "mindhit_redis_cache_operations_total",
                Help: "Total number of Redis cache operations",
            },
            []string{"operation", "result"}, // operation: get/set/delete, result: hit/miss/success/error
        )

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
        AuthAttempts = promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "mindhit_auth_attempts_total",
                Help: "Total number of authentication attempts",
            },
            []string{"method", "status"}, // method: password/google/refresh, status: success/failed
        )

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
        SubscriptionsByPlan = promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "mindhit_subscriptions_by_plan",
                Help: "Number of active subscriptions by plan",
            },
            []string{"plan"}, // "free", "pro", "enterprise"
        )

        TokenUsageDaily = promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "mindhit_token_usage_daily_total",
                Help: "Daily token usage by user plan",
            },
            []string{"plan"},
        )

        UsageLimitExceeded = promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "mindhit_usage_limit_exceeded_total",
                Help: "Number of times usage limit was exceeded",
            },
            []string{"plan", "limit_type"}, // limit_type: daily/monthly
        )
    )
    ```

- [x] **서비스에 메트릭 호출 추가**
  - [x] `internal/service/session_service.go` - 세션 생성/완료 시 메트릭 기록
  - [x] `internal/service/event_service.go` - 이벤트 수신 시 메트릭 기록
  - [x] `internal/service/auth_service.go` - 인증 시도 시 메트릭 기록
  - [ ] `internal/infrastructure/ai/manager.go` - AI 요청 시 메트릭 기록 (TODO: AI 메트릭)
  - [x] `internal/worker/handler/*.go` - Worker 작업 시 메트릭 기록

- [x] **Worker 핸들러에 메트릭 적용 예시**

    ```go
    // internal/worker/handler/mindmap.go
    func (h *MindmapHandler) Handle(ctx context.Context, task *asynq.Task) error {
        start := time.Now()
        jobType := "mindmap_generation"

        defer func() {
            duration := time.Since(start).Seconds()
            metrics.WorkerJobDuration.WithLabelValues(jobType).Observe(duration)
        }()

        // ... processing logic ...

        if err != nil {
            metrics.WorkerJobsProcessed.WithLabelValues(jobType, "failed").Inc()
            return err
        }

        metrics.WorkerJobsProcessed.WithLabelValues(jobType, "success").Inc()
        return nil
    }
    ```

- [ ] **AI Provider에 메트릭 적용 예시** (TODO: 추후 구현)

    ```go
    // internal/infrastructure/ai/manager.go
    func (m *ProviderManager) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
        start := time.Now()
        provider := m.getCurrentProvider()

        resp, err := provider.Chat(ctx, req)

        duration := time.Since(start).Seconds()
        operation := req.Operation // "tag_extraction" or "mindmap_generation"

        if err != nil {
            metrics.AIRequestsTotal.WithLabelValues(provider.Name(), operation, "error").Inc()
            metrics.AIProcessingErrors.WithLabelValues(provider.Name(), operation, "api_error").Inc()
            return nil, err
        }

        metrics.AIRequestsTotal.WithLabelValues(provider.Name(), operation, "success").Inc()
        metrics.AIProcessingDuration.WithLabelValues(provider.Name(), operation).Observe(duration)
        metrics.AITokensUsed.WithLabelValues(provider.Name(), "input").Add(float64(resp.Usage.InputTokens))
        metrics.AITokensUsed.WithLabelValues(provider.Name(), "output").Add(float64(resp.Usage.OutputTokens))

        return resp, nil
    }
    ```

### 검증

```bash
# API 서버 실행 후 메트릭 확인
curl http://localhost:9000/metrics | grep mindhit_

# 특정 메트릭 확인
curl http://localhost:9000/metrics | grep mindhit_ai_
curl http://localhost:9000/metrics | grep mindhit_worker_
```

---

## Step 12.2: Grafana 대시보드 프로비저닝

### 목표

Grafana 대시보드 자동 프로비저닝 및 MindHit 전용 대시보드 구성

### 체크리스트

- [x] **Grafana 프로비저닝 폴더 구조 생성**

    ```bash
    mkdir -p infra/docker/grafana/provisioning/datasources
    mkdir -p infra/docker/grafana/provisioning/dashboards
    mkdir -p infra/docker/grafana/dashboards
    ```

- [x] **docker-compose.yml 업데이트**
  - [x] `infra/docker/docker-compose.yml` Grafana 볼륨 추가

    ```yaml
    grafana:
      image: grafana/grafana:latest
      platform: linux/amd64
      container_name: mindhit-grafana
      ports:
        - "3010:3010"
      volumes:
        - grafana_data:/var/lib/grafana
        - ./grafana/provisioning:/etc/grafana/provisioning
        - ./grafana/dashboards:/var/lib/grafana/dashboards
      environment:
        - GF_SERVER_HTTP_PORT=3010
        - GF_SECURITY_ADMIN_USER=admin
        - GF_SECURITY_ADMIN_PASSWORD=admin
        - GF_USERS_ALLOW_SIGN_UP=false
      depends_on:
        - prometheus
    ```

- [x] **데이터소스 자동 설정**
  - [x] `infra/docker/grafana/provisioning/datasources/datasources.yml`

    ```yaml
    apiVersion: 1

    datasources:
      - name: Prometheus
        type: prometheus
        access: proxy
        url: http://prometheus:9090
        isDefault: true
        editable: false
    ```

- [x] **대시보드 프로비저닝 설정**
  - [x] `infra/docker/grafana/provisioning/dashboards/dashboards.yml`

    ```yaml
    apiVersion: 1

    providers:
      - name: 'MindHit'
        orgId: 1
        folder: 'MindHit'
        type: file
        disableDeletion: false
        updateIntervalSeconds: 30
        options:
          path: /var/lib/grafana/dashboards
    ```

- [x] **API Overview 대시보드**
  - [x] `infra/docker/grafana/dashboards/api-overview.json`

    ```json
    {
      "annotations": {
        "list": []
      },
      "editable": true,
      "fiscalYearStartMonth": 0,
      "graphTooltip": 0,
      "id": null,
      "links": [],
      "liveNow": false,
      "panels": [
        {
          "datasource": { "type": "prometheus", "uid": "prometheus" },
          "fieldConfig": {
            "defaults": { "unit": "reqps" }
          },
          "gridPos": { "h": 8, "w": 12, "x": 0, "y": 0 },
          "id": 1,
          "title": "Request Rate",
          "type": "timeseries",
          "targets": [
            {
              "expr": "sum(rate(mindhit_http_requests_total[5m])) by (method)",
              "legendFormat": "{{method}}"
            }
          ]
        },
        {
          "datasource": { "type": "prometheus", "uid": "prometheus" },
          "fieldConfig": {
            "defaults": { "unit": "s" }
          },
          "gridPos": { "h": 8, "w": 12, "x": 12, "y": 0 },
          "id": 2,
          "title": "Response Time (p95)",
          "type": "timeseries",
          "targets": [
            {
              "expr": "histogram_quantile(0.95, sum(rate(mindhit_http_request_duration_seconds_bucket[5m])) by (le))",
              "legendFormat": "p95"
            },
            {
              "expr": "histogram_quantile(0.50, sum(rate(mindhit_http_request_duration_seconds_bucket[5m])) by (le))",
              "legendFormat": "p50"
            }
          ]
        },
        {
          "datasource": { "type": "prometheus", "uid": "prometheus" },
          "fieldConfig": {
            "defaults": { "unit": "percentunit" }
          },
          "gridPos": { "h": 8, "w": 12, "x": 0, "y": 8 },
          "id": 3,
          "title": "Error Rate",
          "type": "timeseries",
          "targets": [
            {
              "expr": "sum(rate(mindhit_http_requests_total{status=~\"5..\"}[5m])) / sum(rate(mindhit_http_requests_total[5m]))",
              "legendFormat": "5xx Error Rate"
            },
            {
              "expr": "sum(rate(mindhit_http_requests_total{status=~\"4..\"}[5m])) / sum(rate(mindhit_http_requests_total[5m]))",
              "legendFormat": "4xx Error Rate"
            }
          ]
        },
        {
          "datasource": { "type": "prometheus", "uid": "prometheus" },
          "gridPos": { "h": 8, "w": 12, "x": 12, "y": 8 },
          "id": 4,
          "title": "Active Requests",
          "type": "stat",
          "targets": [
            {
              "expr": "mindhit_http_requests_in_flight",
              "legendFormat": "In Flight"
            }
          ]
        }
      ],
      "refresh": "10s",
      "schemaVersion": 38,
      "style": "dark",
      "tags": ["mindhit", "api"],
      "templating": { "list": [] },
      "time": { "from": "now-1h", "to": "now" },
      "timepicker": {},
      "timezone": "",
      "title": "MindHit API Overview",
      "uid": "mindhit-api-overview",
      "version": 1,
      "weekStart": ""
    }
    ```

- [x] **비즈니스 메트릭 대시보드**
  - [x] `infra/docker/grafana/dashboards/business-metrics.json`

    ```json
    {
      "annotations": { "list": [] },
      "editable": true,
      "panels": [
        {
          "datasource": { "type": "prometheus", "uid": "prometheus" },
          "gridPos": { "h": 4, "w": 6, "x": 0, "y": 0 },
          "id": 1,
          "title": "Active Sessions",
          "type": "stat",
          "targets": [
            { "expr": "mindhit_sessions_active" }
          ]
        },
        {
          "datasource": { "type": "prometheus", "uid": "prometheus" },
          "gridPos": { "h": 4, "w": 6, "x": 6, "y": 0 },
          "id": 2,
          "title": "Sessions Created (24h)",
          "type": "stat",
          "targets": [
            { "expr": "increase(mindhit_sessions_created_total[24h])" }
          ]
        },
        {
          "datasource": { "type": "prometheus", "uid": "prometheus" },
          "gridPos": { "h": 4, "w": 6, "x": 12, "y": 0 },
          "id": 3,
          "title": "Events Processed (1h)",
          "type": "stat",
          "targets": [
            { "expr": "increase(mindhit_events_processed_total[1h])" }
          ]
        },
        {
          "datasource": { "type": "prometheus", "uid": "prometheus" },
          "gridPos": { "h": 4, "w": 6, "x": 18, "y": 0 },
          "id": 4,
          "title": "Mindmaps Generated (24h)",
          "type": "stat",
          "targets": [
            { "expr": "increase(mindhit_mindmaps_generated_total{status=\"success\"}[24h])" }
          ]
        },
        {
          "datasource": { "type": "prometheus", "uid": "prometheus" },
          "gridPos": { "h": 8, "w": 12, "x": 0, "y": 4 },
          "id": 5,
          "title": "Events by Type",
          "type": "piechart",
          "targets": [
            {
              "expr": "increase(mindhit_events_received_total[1h])",
              "legendFormat": "{{event_type}}"
            }
          ]
        },
        {
          "datasource": { "type": "prometheus", "uid": "prometheus" },
          "gridPos": { "h": 8, "w": 12, "x": 12, "y": 4 },
          "id": 6,
          "title": "Session Duration Distribution",
          "type": "histogram",
          "targets": [
            { "expr": "mindhit_session_duration_seconds_bucket" }
          ]
        }
      ],
      "refresh": "30s",
      "schemaVersion": 38,
      "tags": ["mindhit", "business"],
      "time": { "from": "now-24h", "to": "now" },
      "title": "MindHit Business Metrics",
      "uid": "mindhit-business",
      "version": 1
    }
    ```

- [x] **AI & Worker 대시보드**
  - [x] `infra/docker/grafana/dashboards/ai-worker.json`

    ```json
    {
      "annotations": { "list": [] },
      "editable": true,
      "panels": [
        {
          "datasource": { "type": "prometheus", "uid": "prometheus" },
          "gridPos": { "h": 8, "w": 12, "x": 0, "y": 0 },
          "id": 1,
          "title": "AI Requests by Provider",
          "type": "timeseries",
          "targets": [
            {
              "expr": "sum(rate(mindhit_ai_requests_total[5m])) by (provider)",
              "legendFormat": "{{provider}}"
            }
          ]
        },
        {
          "datasource": { "type": "prometheus", "uid": "prometheus" },
          "fieldConfig": { "defaults": { "unit": "s" } },
          "gridPos": { "h": 8, "w": 12, "x": 12, "y": 0 },
          "id": 2,
          "title": "AI Processing Time (p95)",
          "type": "timeseries",
          "targets": [
            {
              "expr": "histogram_quantile(0.95, sum(rate(mindhit_ai_processing_duration_seconds_bucket[5m])) by (le, provider))",
              "legendFormat": "{{provider}}"
            }
          ]
        },
        {
          "datasource": { "type": "prometheus", "uid": "prometheus" },
          "gridPos": { "h": 8, "w": 8, "x": 0, "y": 8 },
          "id": 3,
          "title": "AI Tokens Used (1h)",
          "type": "stat",
          "targets": [
            {
              "expr": "sum(increase(mindhit_ai_tokens_used_total[1h])) by (token_type)",
              "legendFormat": "{{token_type}}"
            }
          ]
        },
        {
          "datasource": { "type": "prometheus", "uid": "prometheus" },
          "gridPos": { "h": 8, "w": 8, "x": 8, "y": 8 },
          "id": 4,
          "title": "AI Error Rate",
          "type": "stat",
          "fieldConfig": { "defaults": { "unit": "percentunit" } },
          "targets": [
            {
              "expr": "sum(rate(mindhit_ai_requests_total{status=\"error\"}[5m])) / sum(rate(mindhit_ai_requests_total[5m]))"
            }
          ]
        },
        {
          "datasource": { "type": "prometheus", "uid": "prometheus" },
          "gridPos": { "h": 8, "w": 8, "x": 16, "y": 8 },
          "id": 5,
          "title": "Worker Jobs in Queue",
          "type": "stat",
          "targets": [
            {
              "expr": "sum(mindhit_worker_jobs_in_queue) by (queue)",
              "legendFormat": "{{queue}}"
            }
          ]
        },
        {
          "datasource": { "type": "prometheus", "uid": "prometheus" },
          "gridPos": { "h": 8, "w": 12, "x": 0, "y": 16 },
          "id": 6,
          "title": "Worker Job Processing Rate",
          "type": "timeseries",
          "targets": [
            {
              "expr": "sum(rate(mindhit_worker_jobs_processed_total[5m])) by (job_type, status)",
              "legendFormat": "{{job_type}} - {{status}}"
            }
          ]
        },
        {
          "datasource": { "type": "prometheus", "uid": "prometheus" },
          "fieldConfig": { "defaults": { "unit": "s" } },
          "gridPos": { "h": 8, "w": 12, "x": 12, "y": 16 },
          "id": 7,
          "title": "Worker Job Duration (p95)",
          "type": "timeseries",
          "targets": [
            {
              "expr": "histogram_quantile(0.95, sum(rate(mindhit_worker_job_duration_seconds_bucket[5m])) by (le, job_type))",
              "legendFormat": "{{job_type}}"
            }
          ]
        }
      ],
      "refresh": "30s",
      "schemaVersion": 38,
      "tags": ["mindhit", "ai", "worker"],
      "time": { "from": "now-1h", "to": "now" },
      "title": "MindHit AI & Worker",
      "uid": "mindhit-ai-worker",
      "version": 1
    }
    ```

- [x] **인프라 대시보드**
  - [x] `infra/docker/grafana/dashboards/infrastructure.json`

    ```json
    {
      "annotations": { "list": [] },
      "editable": true,
      "panels": [
        {
          "datasource": { "type": "prometheus", "uid": "prometheus" },
          "fieldConfig": { "defaults": { "unit": "s" } },
          "gridPos": { "h": 8, "w": 12, "x": 0, "y": 0 },
          "id": 1,
          "title": "Database Query Time (p95)",
          "type": "timeseries",
          "targets": [
            {
              "expr": "histogram_quantile(0.95, sum(rate(mindhit_db_query_duration_seconds_bucket[5m])) by (le, operation))",
              "legendFormat": "{{operation}}"
            }
          ]
        },
        {
          "datasource": { "type": "prometheus", "uid": "prometheus" },
          "gridPos": { "h": 8, "w": 12, "x": 12, "y": 0 },
          "id": 2,
          "title": "Database Connections",
          "type": "timeseries",
          "targets": [
            { "expr": "mindhit_db_connections_active", "legendFormat": "Active" },
            { "expr": "mindhit_db_connections_idle", "legendFormat": "Idle" }
          ]
        },
        {
          "datasource": { "type": "prometheus", "uid": "prometheus" },
          "gridPos": { "h": 8, "w": 12, "x": 0, "y": 8 },
          "id": 3,
          "title": "Redis Cache Hit Rate",
          "type": "stat",
          "fieldConfig": { "defaults": { "unit": "percentunit" } },
          "targets": [
            {
              "expr": "sum(rate(mindhit_redis_cache_operations_total{result=\"hit\"}[5m])) / sum(rate(mindhit_redis_cache_operations_total{operation=\"get\"}[5m]))"
            }
          ]
        },
        {
          "datasource": { "type": "prometheus", "uid": "prometheus" },
          "fieldConfig": { "defaults": { "unit": "s" } },
          "gridPos": { "h": 8, "w": 12, "x": 12, "y": 8 },
          "id": 4,
          "title": "Redis Operation Time",
          "type": "timeseries",
          "targets": [
            {
              "expr": "histogram_quantile(0.95, sum(rate(mindhit_redis_operation_duration_seconds_bucket[5m])) by (le, operation))",
              "legendFormat": "{{operation}}"
            }
          ]
        },
        {
          "datasource": { "type": "prometheus", "uid": "prometheus" },
          "gridPos": { "h": 8, "w": 12, "x": 0, "y": 16 },
          "id": 5,
          "title": "Auth Success Rate",
          "type": "stat",
          "fieldConfig": { "defaults": { "unit": "percentunit" } },
          "targets": [
            {
              "expr": "sum(rate(mindhit_auth_attempts_total{status=\"success\"}[1h])) / sum(rate(mindhit_auth_attempts_total[1h]))"
            }
          ]
        },
        {
          "datasource": { "type": "prometheus", "uid": "prometheus" },
          "gridPos": { "h": 8, "w": 12, "x": 12, "y": 16 },
          "id": 6,
          "title": "Auth Attempts by Method",
          "type": "timeseries",
          "targets": [
            {
              "expr": "sum(rate(mindhit_auth_attempts_total[5m])) by (method, status)",
              "legendFormat": "{{method}} - {{status}}"
            }
          ]
        }
      ],
      "refresh": "30s",
      "schemaVersion": 38,
      "tags": ["mindhit", "infrastructure"],
      "time": { "from": "now-1h", "to": "now" },
      "title": "MindHit Infrastructure",
      "uid": "mindhit-infra",
      "version": 1
    }
    ```

### 검증

```bash
# Docker Compose 재시작
cd infra/docker && docker-compose down && docker-compose up -d

# Grafana 접속
# http://localhost:3010
# admin / admin 로그인

# MindHit 폴더에서 대시보드 확인
```

---

## Step 12.3: 로그 수집 시스템 (Loki)

### 목표

Loki를 통한 중앙 집중 로그 수집 및 Grafana 연동

> **Note**: 로컬 개발에서는 `docker logs` 또는 터미널 출력으로 충분합니다.
> Loki는 프로덕션/스테이징 환경에서 더 유용하며, 선택적으로 구성할 수 있습니다.

### 체크리스트

- [ ] **logger.go 컨텍스트 지원 확장** (선택, 추후 구현)
  - [ ] `internal/infrastructure/logger/logger.go` 업데이트

    ```go
    package logger

    import (
        "context"
        "log/slog"
        "os"
    )

    type contextKey string

    const (
        RequestIDKey contextKey = "request_id"
        UserIDKey    contextKey = "user_id"
        SessionIDKey contextKey = "session_id"
    )

    // Init initializes the default slog logger based on the environment.
    func Init(env string) {
        var handler slog.Handler

        opts := &slog.HandlerOptions{
            AddSource: true,
        }

        switch env {
        case "production":
            opts.Level = slog.LevelInfo
            handler = slog.NewJSONHandler(os.Stdout, opts)
        default:
            opts.Level = slog.LevelDebug
            handler = slog.NewTextHandler(os.Stdout, opts)
        }

        slog.SetDefault(slog.New(handler))
    }

    // FromContext creates a logger with context values.
    func FromContext(ctx context.Context) *slog.Logger {
        logger := slog.Default()

        if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
            logger = logger.With("request_id", requestID)
        }
        if userID, ok := ctx.Value(UserIDKey).(string); ok {
            logger = logger.With("user_id", userID)
        }
        if sessionID, ok := ctx.Value(SessionIDKey).(string); ok {
            logger = logger.With("session_id", sessionID)
        }

        return logger
    }

    // WithRequestID adds request ID to context.
    func WithRequestID(ctx context.Context, requestID string) context.Context {
        return context.WithValue(ctx, RequestIDKey, requestID)
    }

    // WithUserID adds user ID to context.
    func WithUserID(ctx context.Context, userID string) context.Context {
        return context.WithValue(ctx, UserIDKey, userID)
    }

    // WithSessionID adds session ID to context.
    func WithSessionID(ctx context.Context, sessionID string) context.Context {
        return context.WithValue(ctx, SessionIDKey, sessionID)
    }
    ```

- [x] **docker-compose.yml에 Loki 추가**
  - [x] `infra/docker/docker-compose.yml`

    ```yaml
    # Loki - Log aggregation (선택적)
    loki:
      image: grafana/loki:2.9.0
      platform: linux/amd64
      container_name: mindhit-loki
      ports:
        - "3100:3100"
      command: -config.file=/etc/loki/local-config.yaml
      volumes:
        - ./loki/loki-config.yaml:/etc/loki/local-config.yaml
        - loki_data:/loki

    # Promtail - Log collector (선택적)
    promtail:
      image: grafana/promtail:2.9.0
      platform: linux/amd64
      container_name: mindhit-promtail
      volumes:
        - ./promtail/promtail-config.yaml:/etc/promtail/config.yaml
        - /var/log:/var/log:ro
        - /var/lib/docker/containers:/var/lib/docker/containers:ro
      command: -config.file=/etc/promtail/config.yaml
      depends_on:
        - loki
    ```

- [x] **Loki 설정 파일**
  - [x] `infra/docker/loki/loki-config.yaml`

    ```yaml
    auth_enabled: false

    server:
      http_listen_port: 3100

    common:
      path_prefix: /loki
      storage:
        filesystem:
          chunks_directory: /loki/chunks
          rules_directory: /loki/rules
      replication_factor: 1
      ring:
        kvstore:
          store: inmemory

    schema_config:
      configs:
        - from: 2020-10-24
          store: boltdb-shipper
          object_store: filesystem
          schema: v11
          index:
            prefix: index_
            period: 24h
    ```

- [x] **Promtail 설정 파일**
  - [x] `infra/docker/promtail/promtail-config.yaml`

    ```yaml
    server:
      http_listen_port: 9080
      grpc_listen_port: 0

    positions:
      filename: /tmp/positions.yaml

    clients:
      - url: http://loki:3100/loki/api/v1/push

    scrape_configs:
      - job_name: docker
        static_configs:
          - targets:
              - localhost
            labels:
              job: containerlogs
              __path__: /var/lib/docker/containers/*/*-json.log
        pipeline_stages:
          - json:
              expressions:
                output: log
                stream: stream
                time: time
          - json:
              expressions:
                level: level
                msg: msg
                request_id: request_id
                user_id: user_id
              source: output
          - labels:
              level:
              request_id:
          - output:
              source: output
    ```

- [x] **Grafana 데이터소스에 Loki 추가**
  - [x] `infra/docker/grafana/provisioning/datasources/datasources.yml` 업데이트

    ```yaml
    apiVersion: 1

    datasources:
      - name: Prometheus
        type: prometheus
        access: proxy
        url: http://prometheus:9090
        isDefault: true
        editable: false

      - name: Loki
        type: loki
        access: proxy
        url: http://loki:3100
        editable: false
    ```

### 검증

```bash
# Loki 헬스 체크
curl http://localhost:3100/ready

# Grafana에서 Loki 로그 쿼리
# Explore > Loki 데이터소스 선택
# {job="containerlogs"} |= "mindhit"
```

---

## Step 12.4: 알림 시스템 구성

### 목표

Alertmanager를 통한 알림 시스템 구축 (Slack/Email 연동)

### 체크리스트

- [x] **prometheus.yml에 Alertmanager 연동 추가**
  - [x] `infra/docker/prometheus.yml` 업데이트

    ```yaml
    global:
      scrape_interval: 15s
      evaluation_interval: 15s

    alerting:
      alertmanagers:
        - static_configs:
            - targets:
              - alertmanager:9093

    rule_files:
      - /etc/prometheus/alerts.yml

    scrape_configs:
      - job_name: "prometheus"
        static_configs:
          - targets: ["localhost:9090"]

      - job_name: "mindhit-api"
        static_configs:
          - targets: ["host.docker.internal:9000"]
        metrics_path: /metrics
    ```

- [x] **알림 규칙 파일 생성**
  - [x] `infra/docker/prometheus/alerts.yml`

    ```yaml
    groups:
      - name: mindhit-api
        rules:
          # API 서버 다운
          - alert: APIDown
            expr: up{job="mindhit-api"} == 0
            for: 1m
            labels:
              severity: critical
            annotations:
              summary: "MindHit API server is down"
              description: "API server has been down for more than 1 minute"

          # 높은 에러율 (5xx > 1%)
          - alert: HighErrorRate
            expr: |
              (
                sum(rate(mindhit_http_requests_total{status=~"5.."}[5m]))
                /
                sum(rate(mindhit_http_requests_total[5m]))
              ) > 0.01
            for: 5m
            labels:
              severity: critical
            annotations:
              summary: "High API error rate"
              description: "Error rate is {{ $value | humanizePercentage }} (threshold: 1%)"

          # 높은 지연시간 (p95 > 2s)
          - alert: HighLatency
            expr: |
              histogram_quantile(0.95, sum(rate(mindhit_http_request_duration_seconds_bucket[5m])) by (le)) > 2
            for: 5m
            labels:
              severity: warning
            annotations:
              summary: "High API latency"
              description: "95th percentile latency is {{ $value }}s (threshold: 2s)"

      - name: mindhit-ai
        rules:
          # AI 처리 에러 급증
          - alert: AIProcessingErrors
            expr: increase(mindhit_ai_processing_errors_total[1h]) > 10
            for: 15m
            labels:
              severity: warning
            annotations:
              summary: "High AI processing errors"
              description: "{{ $value }} AI processing errors in the last hour"

          # AI 응답 시간 느림
          - alert: SlowAIProcessing
            expr: |
              histogram_quantile(0.95, sum(rate(mindhit_ai_processing_duration_seconds_bucket[5m])) by (le)) > 60
            for: 10m
            labels:
              severity: warning
            annotations:
              summary: "Slow AI processing"
              description: "AI processing p95 is {{ $value }}s (threshold: 60s)"

      - name: mindhit-worker
        rules:
          # Worker 작업 실패율 높음
          - alert: HighWorkerFailureRate
            expr: |
              (
                sum(rate(mindhit_worker_jobs_processed_total{status="failed"}[1h]))
                /
                sum(rate(mindhit_worker_jobs_processed_total[1h]))
              ) > 0.1
            for: 15m
            labels:
              severity: warning
            annotations:
              summary: "High worker job failure rate"
              description: "Worker failure rate is {{ $value | humanizePercentage }}"

          # 대기 중인 작업 너무 많음
          - alert: HighQueueBacklog
            expr: sum(mindhit_worker_jobs_in_queue{state="pending"}) > 100
            for: 15m
            labels:
              severity: warning
            annotations:
              summary: "High queue backlog"
              description: "{{ $value }} jobs pending in queue"

      - name: mindhit-infrastructure
        rules:
          # DB 쿼리 느림
          - alert: SlowDBQueries
            expr: |
              histogram_quantile(0.95, sum(rate(mindhit_db_query_duration_seconds_bucket[5m])) by (le)) > 0.5
            for: 10m
            labels:
              severity: warning
            annotations:
              summary: "Slow database queries"
              description: "DB query p95 is {{ $value }}s (threshold: 0.5s)"

          # Redis 캐시 히트율 낮음
          - alert: LowCacheHitRate
            expr: |
              (
                sum(rate(mindhit_redis_cache_operations_total{result="hit"}[5m]))
                /
                sum(rate(mindhit_redis_cache_operations_total{operation="get"}[5m]))
              ) < 0.5
            for: 30m
            labels:
              severity: warning
            annotations:
              summary: "Low cache hit rate"
              description: "Cache hit rate is {{ $value | humanizePercentage }}"
    ```

- [x] **docker-compose.yml에 Alertmanager 추가**
  - [x] `infra/docker/docker-compose.yml`

    ```yaml
    # Alertmanager - Alert management
    alertmanager:
      image: prom/alertmanager:v0.26.0
      platform: linux/amd64
      container_name: mindhit-alertmanager
      ports:
        - "9093:9093"
      volumes:
        - ./alertmanager/alertmanager.yml:/etc/alertmanager/alertmanager.yml
        - alertmanager_data:/alertmanager
      command:
        - '--config.file=/etc/alertmanager/alertmanager.yml'
        - '--storage.path=/alertmanager'
    ```

- [x] **Alertmanager 설정 파일**
  - [x] `infra/docker/alertmanager/alertmanager.yml`

    ```yaml
    global:
      resolve_timeout: 5m
      # Slack webhook URL (환경변수로 설정)
      # slack_api_url: '${SLACK_WEBHOOK_URL}'

    route:
      group_by: ['alertname', 'severity']
      group_wait: 30s
      group_interval: 5m
      repeat_interval: 4h
      receiver: 'default'
      routes:
        - match:
            severity: critical
          receiver: 'critical'
          continue: true
        - match:
            severity: warning
          receiver: 'default'

    receivers:
      - name: 'default'
        # Slack 설정 (선택)
        # slack_configs:
        #   - channel: '#mindhit-alerts'
        #     send_resolved: true
        #     title: '{{ if eq .Status "firing" }}🔥{{ else }}✅{{ end }} {{ .CommonAnnotations.summary }}'
        #     text: '{{ .CommonAnnotations.description }}'

        # Webhook 설정 (선택)
        webhook_configs:
          - url: 'http://host.docker.internal:9000/webhooks/alerts'
            send_resolved: true

      - name: 'critical'
        # Critical 알림용 (예: PagerDuty, Slack critical 채널)
        webhook_configs:
          - url: 'http://host.docker.internal:9000/webhooks/alerts/critical'
            send_resolved: true

    inhibit_rules:
      - source_match:
          severity: 'critical'
        target_match:
          severity: 'warning'
        equal: ['alertname']
    ```

- [x] **Prometheus 볼륨 업데이트**
  - [x] `infra/docker/docker-compose.yml` prometheus 서비스 업데이트

    ```yaml
    prometheus:
      image: prom/prometheus:latest
      platform: linux/amd64
      container_name: mindhit-prometheus
      ports:
        - "9091:9090"
      volumes:
        - ./prometheus.yml:/etc/prometheus/prometheus.yml
        - ./prometheus/alerts.yml:/etc/prometheus/alerts.yml
        - prometheus_data:/prometheus
      command:
        - "--config.file=/etc/prometheus/prometheus.yml"
        - "--storage.tsdb.path=/prometheus"
        - "--web.enable-lifecycle"
      extra_hosts:
        - "host.docker.internal:host-gateway"
    ```

- [x] **환경 변수 파일 업데이트**
  - [x] `.env.example`에 추가

    ```env
    # Alerting (선택)
    SLACK_WEBHOOK_URL=https://hooks.slack.com/services/xxx
    ```

### 검증

```bash
# Alertmanager 헬스 체크
curl http://localhost:9093/-/healthy

# Prometheus Alerts 확인
# http://localhost:9091/alerts

# Alertmanager UI 확인
# http://localhost:9093

# 알림 규칙 검증
docker exec mindhit-prometheus promtool check rules /etc/prometheus/alerts.yml
```

---

## Phase 12 완료 확인

### 전체 검증 체크리스트

- [ ] **메트릭 수집**
  - [ ] `/metrics` 엔드포인트 응답
  - [ ] HTTP 메트릭 수집 확인 (`mindhit_http_*`)
  - [ ] 비즈니스 메트릭 수집 확인 (`mindhit_sessions_*`, `mindhit_events_*`)
  - [ ] AI 메트릭 수집 확인 (`mindhit_ai_*`)
  - [ ] Worker 메트릭 수집 확인 (`mindhit_worker_*`)

- [ ] **Grafana 대시보드**
  - [ ] 로그인 가능 (`http://localhost:3010`)
  - [ ] API Overview 대시보드 표시
  - [ ] Business Metrics 대시보드 표시
  - [ ] AI & Worker 대시보드 표시
  - [ ] Infrastructure 대시보드 표시

- [ ] **로깅 시스템** (선택)
  - [ ] 구조화된 JSON 로그 출력 (production 환경)
  - [ ] Request ID 추적
  - [ ] Loki에서 로그 검색

- [ ] **알림 시스템**
  - [ ] Prometheus 알림 규칙 로드됨
  - [ ] Alertmanager 실행 중
  - [ ] 테스트 알림 수신

### 테스트 요구사항

| 테스트 유형 | 대상 | 검증 방법 |
| ----------- | ---- | --------- |
| 통합 테스트 | 메트릭 수집 | Prometheus API 쿼리 |
| 통합 테스트 | 로그 수집 | Loki API 쿼리 (선택) |
| 알림 테스트 | 알림 규칙 | `promtool check rules` |
| 회귀 테스트 | 기존 테스트 통과 | `moonx backend:test` |

```bash
# Phase 12 검증
# 1. 전체 테스트 통과 확인
moonx backend:test

# 2. 모니터링 스택 헬스 체크
curl http://localhost:9091/-/healthy  # Prometheus
curl http://localhost:3010/api/health # Grafana
curl http://localhost:9093/-/healthy  # Alertmanager

# 3. 메트릭 확인
curl http://localhost:9000/metrics | grep mindhit_
```

### 산출물 요약

| 항목 | 위치 |
| ---- | ---- |
| 비즈니스 메트릭 | `internal/infrastructure/metrics/metrics.go` |
| Grafana 데이터소스 | `infra/docker/grafana/provisioning/datasources/datasources.yml` |
| Grafana 대시보드 | `infra/docker/grafana/dashboards/*.json` |
| Prometheus 설정 | `infra/docker/prometheus.yml` |
| 알림 규칙 | `infra/docker/prometheus/alerts.yml` |
| Alertmanager 설정 | `infra/docker/alertmanager/alertmanager.yml` |
| Loki 설정 (선택) | `infra/docker/loki/loki-config.yaml` |
| Promtail 설정 (선택) | `infra/docker/promtail/promtail-config.yaml` |

### 모니터링 스택 요약

| 서비스 | 포트 | 용도 |
|-------|------|------|
| Prometheus | 9091 | 메트릭 수집/저장 |
| Grafana | 3010 | 대시보드/시각화 |
| Alertmanager | 9093 | 알림 관리 |
| Loki | 3100 | 로그 집계 (선택) |
| Asynqmon | 9090 | Asynq 작업 모니터링 |

### 핵심 메트릭

| 카테고리 | 메트릭 | 용도 |
|---------|-------|------|
| HTTP | `mindhit_http_requests_total` | API 요청 수 |
| HTTP | `mindhit_http_request_duration_seconds` | 응답 시간 |
| Session | `mindhit_sessions_active` | 활성 세션 수 |
| Session | `mindhit_sessions_created_total` | 생성된 세션 수 |
| Event | `mindhit_events_received_total` | 이벤트 수신 수 |
| AI | `mindhit_ai_requests_total` | AI 요청 수 |
| AI | `mindhit_ai_processing_duration_seconds` | AI 처리 시간 |
| AI | `mindhit_ai_tokens_used_total` | 토큰 사용량 |
| Worker | `mindhit_worker_jobs_processed_total` | 처리된 작업 수 |
| Worker | `mindhit_worker_jobs_in_queue` | 대기 중인 작업 |
| DB | `mindhit_db_query_duration_seconds` | DB 쿼리 시간 |
| Cache | `mindhit_redis_cache_operations_total` | 캐시 작업 수 |

---

## 폴더 구조

Phase 12 완료 후 infra/docker 폴더 구조:

```text
infra/docker/
├── docker-compose.yml
├── prometheus.yml
├── prometheus/
│   └── alerts.yml
├── grafana/
│   ├── provisioning/
│   │   ├── datasources/
│   │   │   └── datasources.yml
│   │   └── dashboards/
│   │       └── dashboards.yml
│   └── dashboards/
│       ├── api-overview.json
│       ├── business-metrics.json
│       ├── ai-worker.json
│       └── infrastructure.json
├── alertmanager/
│   └── alertmanager.yml
├── loki/                    # 선택
│   └── loki-config.yaml
└── promtail/                # 선택
    └── promtail-config.yaml
```

---

## 다음 Phase

Phase 12 완료 후 [Phase 13: 배포/운영](./phase-13-deployment.md)으로 진행하세요.
