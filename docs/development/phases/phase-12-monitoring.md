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

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 12.1 | Prometheus 메트릭 수집 | ⬜ |
| 12.2 | Grafana 대시보드 구성 | ⬜ |
| 12.3 | 구조화된 로깅 시스템 | ⬜ |
| 12.4 | 알림 시스템 구성 | ⬜ |

---

## Step 12.1: Prometheus 메트릭 수집

### 목표

API 서버의 핵심 메트릭 수집 및 Prometheus 엔드포인트 노출

### 체크리스트

- [ ] **Prometheus 클라이언트 의존성 추가**

  ```bash
  cd apps/api
  go get github.com/prometheus/client_golang/prometheus
  go get github.com/prometheus/client_golang/prometheus/promauto
  go get github.com/prometheus/client_golang/prometheus/promhttp
  ```

- [ ] **메트릭 미들웨어 구현**
  - [ ] `internal/infrastructure/middleware/metrics.go`

    ```go
    package middleware

    import (
        "strconv"
        "time"

        "github.com/gin-gonic/gin"
        "github.com/prometheus/client_golang/prometheus"
        "github.com/prometheus/client_golang/prometheus/promauto"
    )

    var (
        // HTTP 요청 총 수
        httpRequestsTotal = promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "http_requests_total",
                Help: "Total number of HTTP requests",
            },
            []string{"method", "path", "status"},
        )

        // HTTP 요청 지속시간
        httpRequestDuration = promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "http_request_duration_seconds",
                Help:    "HTTP request duration in seconds",
                Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
            },
            []string{"method", "path"},
        )

        // 활성 연결 수
        httpActiveConnections = promauto.NewGauge(
            prometheus.GaugeOpts{
                Name: "http_active_connections",
                Help: "Number of active HTTP connections",
            },
        )

        // 요청 크기
        httpRequestSize = promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "http_request_size_bytes",
                Help:    "HTTP request size in bytes",
                Buckets: prometheus.ExponentialBuckets(100, 10, 8),
            },
            []string{"method", "path"},
        )

        // 응답 크기
        httpResponseSize = promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "http_response_size_bytes",
                Help:    "HTTP response size in bytes",
                Buckets: prometheus.ExponentialBuckets(100, 10, 8),
            },
            []string{"method", "path"},
        )
    )

    func Metrics() gin.HandlerFunc {
        return func(c *gin.Context) {
            start := time.Now()
            path := c.FullPath()
            if path == "" {
                path = "unknown"
            }

            httpActiveConnections.Inc()
            defer httpActiveConnections.Dec()

            // 요청 크기 기록
            httpRequestSize.WithLabelValues(c.Request.Method, path).
                Observe(float64(c.Request.ContentLength))

            c.Next()

            // 메트릭 기록
            duration := time.Since(start).Seconds()
            status := strconv.Itoa(c.Writer.Status())

            httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
            httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
            httpResponseSize.WithLabelValues(c.Request.Method, path).
                Observe(float64(c.Writer.Size()))
        }
    }
    ```

- [ ] **비즈니스 메트릭 정의**
  - [ ] `internal/infrastructure/metrics/business.go`

    ```go
    package metrics

    import (
        "github.com/prometheus/client_golang/prometheus"
        "github.com/prometheus/client_golang/prometheus/promauto"
    )

    var (
        // 세션 관련 메트릭
        SessionsCreated = promauto.NewCounter(
            prometheus.CounterOpts{
                Name: "mindhit_sessions_created_total",
                Help: "Total number of sessions created",
            },
        )

        SessionsCompleted = promauto.NewCounter(
            prometheus.CounterOpts{
                Name: "mindhit_sessions_completed_total",
                Help: "Total number of sessions completed",
            },
        )

        SessionDuration = promauto.NewHistogram(
            prometheus.HistogramOpts{
                Name:    "mindhit_session_duration_seconds",
                Help:    "Session duration in seconds",
                Buckets: []float64{60, 300, 600, 1800, 3600, 7200},
            },
        )

        ActiveSessions = promauto.NewGauge(
            prometheus.GaugeOpts{
                Name: "mindhit_active_sessions",
                Help: "Number of currently active sessions",
            },
        )

        // 이벤트 관련 메트릭
        EventsReceived = promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "mindhit_events_received_total",
                Help: "Total number of events received",
            },
            []string{"event_type"},
        )

        EventBatchSize = promauto.NewHistogram(
            prometheus.HistogramOpts{
                Name:    "mindhit_event_batch_size",
                Help:    "Number of events per batch",
                Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500},
            },
        )

        // AI 처리 메트릭
        AIProcessingDuration = promauto.NewHistogram(
            prometheus.HistogramOpts{
                Name:    "mindhit_ai_processing_duration_seconds",
                Help:    "AI processing duration in seconds",
                Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
            },
        )

        AIProcessingErrors = promauto.NewCounter(
            prometheus.CounterOpts{
                Name: "mindhit_ai_processing_errors_total",
                Help: "Total number of AI processing errors",
            },
        )

        // 데이터베이스 메트릭
        DBQueryDuration = promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "mindhit_db_query_duration_seconds",
                Help:    "Database query duration in seconds",
                Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
            },
            []string{"operation"},
        )

        DBConnectionsActive = promauto.NewGauge(
            prometheus.GaugeOpts{
                Name: "mindhit_db_connections_active",
                Help: "Number of active database connections",
            },
        )

        // Redis 메트릭
        RedisCacheHits = promauto.NewCounter(
            prometheus.CounterOpts{
                Name: "mindhit_redis_cache_hits_total",
                Help: "Total number of Redis cache hits",
            },
        )

        RedisCacheMisses = promauto.NewCounter(
            prometheus.CounterOpts{
                Name: "mindhit_redis_cache_misses_total",
                Help: "Total number of Redis cache misses",
            },
        )

        // 인증 메트릭
        AuthLoginAttempts = promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "mindhit_auth_login_attempts_total",
                Help: "Total number of login attempts",
            },
            []string{"status"}, // "success", "failed"
        )

        AuthTokenRefreshes = promauto.NewCounter(
            prometheus.CounterOpts{
                Name: "mindhit_auth_token_refreshes_total",
                Help: "Total number of token refreshes",
            },
        )
    )
    ```

- [ ] **main.go에 메트릭 엔드포인트 추가**

  ```go
  import "github.com/prometheus/client_golang/prometheus/promhttp"

  // /metrics 엔드포인트 추가
  r.GET("/metrics", gin.WrapH(promhttp.Handler()))
  ```

- [ ] **Docker Compose에 Prometheus 추가**
  - [ ] `docker-compose.monitoring.yml`

    ```yaml
    version: '3.8'

    services:
      prometheus:
        image: prom/prometheus:v2.48.0
        container_name: mindhit-prometheus
        ports:
          - "9090:9090"
        volumes:
          - ./monitoring/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
          - ./monitoring/prometheus/alerts.yml:/etc/prometheus/alerts.yml
          - prometheus_data:/prometheus
        command:
          - '--config.file=/etc/prometheus/prometheus.yml'
          - '--storage.tsdb.path=/prometheus'
          - '--web.enable-lifecycle'
        networks:
          - mindhit-network

    volumes:
      prometheus_data:

    networks:
      mindhit-network:
        external: true
    ```

- [ ] **Prometheus 설정 파일 생성**
  - [ ] `monitoring/prometheus/prometheus.yml`

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
      - "alerts.yml"

    scrape_configs:
      - job_name: 'prometheus'
        static_configs:
          - targets: ['localhost:9090']

      - job_name: 'mindhit-api'
        static_configs:
          - targets: ['api:8080']
        metrics_path: /metrics
        scheme: http

      - job_name: 'postgres'
        static_configs:
          - targets: ['postgres-exporter:9187']

      - job_name: 'redis'
        static_configs:
          - targets: ['redis-exporter:9121']
    ```

### 검증

```bash
# API 서버 실행 후 메트릭 확인
curl http://localhost:8080/metrics

# Prometheus UI 확인
# http://localhost:9090
```

---

## Step 12.2: Grafana 대시보드 구성

### 목표

Prometheus 데이터를 시각화하는 Grafana 대시보드 구성

### 체크리스트

- [ ] **Docker Compose에 Grafana 추가**
  - [ ] `docker-compose.monitoring.yml`에 추가

    ```yaml
    grafana:
      image: grafana/grafana:10.2.0
      container_name: mindhit-grafana
      ports:
        - "3001:3000"
      environment:
        - GF_SECURITY_ADMIN_USER=admin
        - GF_SECURITY_ADMIN_PASSWORD=admin
        - GF_USERS_ALLOW_SIGN_UP=false
      volumes:
        - grafana_data:/var/lib/grafana
        - ./monitoring/grafana/provisioning:/etc/grafana/provisioning
        - ./monitoring/grafana/dashboards:/var/lib/grafana/dashboards
      depends_on:
        - prometheus
      networks:
        - mindhit-network
    ```

- [ ] **Grafana 데이터소스 자동 설정**
  - [ ] `monitoring/grafana/provisioning/datasources/datasource.yml`

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

- [ ] **대시보드 자동 프로비저닝**
  - [ ] `monitoring/grafana/provisioning/dashboards/dashboard.yml`

    ```yaml
    apiVersion: 1

    providers:
      - name: 'MindHit Dashboards'
        orgId: 1
        folder: 'MindHit'
        type: file
        disableDeletion: false
        updateIntervalSeconds: 30
        options:
          path: /var/lib/grafana/dashboards
    ```

- [ ] **API 서버 대시보드**
  - [ ] `monitoring/grafana/dashboards/api-overview.json`

    ```json
    {
      "title": "MindHit API Overview",
      "uid": "mindhit-api-overview",
      "panels": [
        {
          "title": "Request Rate",
          "type": "graph",
          "targets": [
            {
              "expr": "rate(http_requests_total[5m])",
              "legendFormat": "{{method}} {{path}}"
            }
          ]
        },
        {
          "title": "Response Time (p95)",
          "type": "graph",
          "targets": [
            {
              "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
              "legendFormat": "p95"
            }
          ]
        },
        {
          "title": "Error Rate",
          "type": "graph",
          "targets": [
            {
              "expr": "rate(http_requests_total{status=~\"5..\"}[5m]) / rate(http_requests_total[5m])",
              "legendFormat": "Error Rate"
            }
          ]
        },
        {
          "title": "Active Connections",
          "type": "gauge",
          "targets": [
            {
              "expr": "http_active_connections",
              "legendFormat": "Active"
            }
          ]
        }
      ]
    }
    ```

- [ ] **세션/비즈니스 메트릭 대시보드**
  - [ ] `monitoring/grafana/dashboards/business-metrics.json`

    ```json
    {
      "title": "MindHit Business Metrics",
      "uid": "mindhit-business",
      "panels": [
        {
          "title": "Active Sessions",
          "type": "stat",
          "targets": [
            {
              "expr": "mindhit_active_sessions"
            }
          ]
        },
        {
          "title": "Sessions Created (24h)",
          "type": "stat",
          "targets": [
            {
              "expr": "increase(mindhit_sessions_created_total[24h])"
            }
          ]
        },
        {
          "title": "Session Duration Distribution",
          "type": "histogram",
          "targets": [
            {
              "expr": "mindhit_session_duration_seconds_bucket"
            }
          ]
        },
        {
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
          "title": "AI Processing Time",
          "type": "graph",
          "targets": [
            {
              "expr": "histogram_quantile(0.95, rate(mindhit_ai_processing_duration_seconds_bucket[5m]))",
              "legendFormat": "p95"
            }
          ]
        }
      ]
    }
    ```

- [ ] **인프라 대시보드**
  - [ ] `monitoring/grafana/dashboards/infrastructure.json`

    ```json
    {
      "title": "MindHit Infrastructure",
      "uid": "mindhit-infra",
      "panels": [
        {
          "title": "Database Query Time (p95)",
          "type": "graph",
          "targets": [
            {
              "expr": "histogram_quantile(0.95, rate(mindhit_db_query_duration_seconds_bucket[5m]))",
              "legendFormat": "{{operation}}"
            }
          ]
        },
        {
          "title": "Database Connections",
          "type": "gauge",
          "targets": [
            {
              "expr": "mindhit_db_connections_active"
            }
          ]
        },
        {
          "title": "Redis Cache Hit Rate",
          "type": "stat",
          "targets": [
            {
              "expr": "rate(mindhit_redis_cache_hits_total[5m]) / (rate(mindhit_redis_cache_hits_total[5m]) + rate(mindhit_redis_cache_misses_total[5m]))"
            }
          ]
        },
        {
          "title": "Auth Login Success Rate",
          "type": "stat",
          "targets": [
            {
              "expr": "rate(mindhit_auth_login_attempts_total{status=\"success\"}[1h]) / rate(mindhit_auth_login_attempts_total[1h])"
            }
          ]
        }
      ]
    }
    ```

### 검증

```bash
# Grafana 접속
# http://localhost:3001
# admin / admin 로그인

# 대시보드 확인
# MindHit 폴더 아래 대시보드들 확인
```

---

## Step 12.3: 구조화된 로깅 시스템

### 목표

JSON 형식의 구조화된 로그와 로그 수집 시스템 구성

### 체크리스트

- [ ] **구조화된 로거 설정**
  - [ ] `internal/infrastructure/logger/logger.go`

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

    // Setup initializes the global logger
    func Setup(environment string) {
        var handler slog.Handler

        opts := &slog.HandlerOptions{
            Level: slog.LevelInfo,
        }

        if environment == "development" {
            opts.Level = slog.LevelDebug
            handler = slog.NewTextHandler(os.Stdout, opts)
        } else {
            handler = slog.NewJSONHandler(os.Stdout, opts)
        }

        logger := slog.New(handler)
        slog.SetDefault(logger)
    }

    // FromContext creates a logger with context values
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

    // WithRequestID adds request ID to context
    func WithRequestID(ctx context.Context, requestID string) context.Context {
        return context.WithValue(ctx, RequestIDKey, requestID)
    }

    // WithUserID adds user ID to context
    func WithUserID(ctx context.Context, userID string) context.Context {
        return context.WithValue(ctx, UserIDKey, userID)
    }

    // WithSessionID adds session ID to context
    func WithSessionID(ctx context.Context, sessionID string) context.Context {
        return context.WithValue(ctx, SessionIDKey, sessionID)
    }
    ```

- [ ] **요청 로깅 미들웨어**
  - [ ] `internal/infrastructure/middleware/logging.go`

    ```go
    package middleware

    import (
        "log/slog"
        "time"

        "github.com/gin-gonic/gin"
        "github.com/google/uuid"

        "github.com/mindhit/api/internal/infrastructure/logger"
    )

    func Logging() gin.HandlerFunc {
        return func(c *gin.Context) {
            start := time.Now()
            requestID := uuid.New().String()

            // 컨텍스트에 request ID 추가
            ctx := logger.WithRequestID(c.Request.Context(), requestID)
            c.Request = c.Request.WithContext(ctx)
            c.Header("X-Request-ID", requestID)

            // 요청 처리
            c.Next()

            // 로그 기록
            duration := time.Since(start)
            status := c.Writer.Status()

            logLevel := slog.LevelInfo
            if status >= 500 {
                logLevel = slog.LevelError
            } else if status >= 400 {
                logLevel = slog.LevelWarn
            }

            slog.Log(c.Request.Context(), logLevel, "request completed",
                "method", c.Request.Method,
                "path", c.Request.URL.Path,
                "status", status,
                "duration_ms", duration.Milliseconds(),
                "client_ip", c.ClientIP(),
                "user_agent", c.Request.UserAgent(),
                "request_id", requestID,
            )
        }
    }
    ```

- [ ] **로그 집계 (Loki) 설정**
  - [ ] `docker-compose.monitoring.yml`에 추가

    ```yaml
    loki:
      image: grafana/loki:2.9.0
      container_name: mindhit-loki
      ports:
        - "3100:3100"
      command: -config.file=/etc/loki/local-config.yaml
      volumes:
        - ./monitoring/loki/loki-config.yaml:/etc/loki/local-config.yaml
        - loki_data:/loki
      networks:
        - mindhit-network

    promtail:
      image: grafana/promtail:2.9.0
      container_name: mindhit-promtail
      volumes:
        - ./monitoring/promtail/promtail-config.yaml:/etc/promtail/config.yaml
        - /var/log:/var/log:ro
        - /var/lib/docker/containers:/var/lib/docker/containers:ro
      command: -config.file=/etc/promtail/config.yaml
      networks:
        - mindhit-network
    ```

- [ ] **Loki 설정**
  - [ ] `monitoring/loki/loki-config.yaml`

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

    ruler:
      alertmanager_url: http://alertmanager:9093
    ```

- [ ] **Promtail 설정**
  - [ ] `monitoring/promtail/promtail-config.yaml`

    ```yaml
    server:
      http_listen_port: 9080
      grpc_listen_port: 0

    positions:
      filename: /tmp/positions.yaml

    clients:
      - url: http://loki:3100/loki/api/v1/push

    scrape_configs:
      - job_name: containers
        static_configs:
          - targets:
              - localhost
            labels:
              job: containerlogs
              __path__: /var/lib/docker/containers/*/*log.json
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
              source: output
          - labels:
              level:
              request_id:
          - output:
              source: output
    ```

- [ ] **Grafana에 Loki 데이터소스 추가**
  - [ ] `monitoring/grafana/provisioning/datasources/datasource.yml`에 추가

    ```yaml
    - name: Loki
      type: loki
      access: proxy
      url: http://loki:3100
      isDefault: false
      editable: false
    ```

### 검증

```bash
# 로그 확인 (콘솔)
docker-compose logs -f api

# Grafana에서 Loki 로그 쿼리
# Explore > Loki
# {job="containerlogs"} |= "mindhit"
```

---

## Step 12.4: 알림 시스템 구성

### 목표

Alertmanager를 통한 알림 시스템 구축

### 체크리스트

- [ ] **Prometheus 알림 규칙 정의**
  - [ ] `monitoring/prometheus/alerts.yml`

    ```yaml
    groups:
      - name: mindhit-api
        rules:
          # API 에러율 높음
          - alert: HighErrorRate
            expr: |
              (
                sum(rate(http_requests_total{status=~"5.."}[5m]))
                /
                sum(rate(http_requests_total[5m]))
              ) > 0.01
            for: 5m
            labels:
              severity: critical
            annotations:
              summary: "High API error rate"
              description: "Error rate is {{ $value | humanizePercentage }} (threshold: 1%)"

          # API 응답 시간 느림
          - alert: HighLatency
            expr: |
              histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
            for: 5m
            labels:
              severity: warning
            annotations:
              summary: "High API latency"
              description: "95th percentile latency is {{ $value }}s"

          # API 서버 다운
          - alert: APIDown
            expr: up{job="mindhit-api"} == 0
            for: 1m
            labels:
              severity: critical
            annotations:
              summary: "API server is down"
              description: "MindHit API server is not responding"

      - name: mindhit-business
        rules:
          # 세션 처리 실패 급증
          - alert: HighSessionFailureRate
            expr: |
              (
                increase(mindhit_sessions_completed_total{status="failed"}[1h])
                /
                increase(mindhit_sessions_completed_total[1h])
              ) > 0.1
            for: 15m
            labels:
              severity: warning
            annotations:
              summary: "High session failure rate"
              description: "Session failure rate is {{ $value | humanizePercentage }}"

          # AI 처리 에러 급증
          - alert: AIProcessingErrors
            expr: increase(mindhit_ai_processing_errors_total[1h]) > 10
            for: 15m
            labels:
              severity: warning
            annotations:
              summary: "High AI processing errors"
              description: "{{ $value }} AI processing errors in the last hour"

      - name: mindhit-infrastructure
        rules:
          # 데이터베이스 연결 부족
          - alert: LowDBConnections
            expr: mindhit_db_connections_active < 2
            for: 5m
            labels:
              severity: warning
            annotations:
              summary: "Low database connections"
              description: "Only {{ $value }} active database connections"

          # 데이터베이스 쿼리 느림
          - alert: SlowDBQueries
            expr: |
              histogram_quantile(0.95, rate(mindhit_db_query_duration_seconds_bucket[5m])) > 0.5
            for: 10m
            labels:
              severity: warning
            annotations:
              summary: "Slow database queries"
              description: "95th percentile query time is {{ $value }}s"

          # Redis 캐시 히트율 낮음
          - alert: LowCacheHitRate
            expr: |
              (
                rate(mindhit_redis_cache_hits_total[5m])
                /
                (rate(mindhit_redis_cache_hits_total[5m]) + rate(mindhit_redis_cache_misses_total[5m]))
              ) < 0.5
            for: 30m
            labels:
              severity: warning
            annotations:
              summary: "Low Redis cache hit rate"
              description: "Cache hit rate is {{ $value | humanizePercentage }}"
    ```

- [ ] **Alertmanager 추가**
  - [ ] `docker-compose.monitoring.yml`에 추가

    ```yaml
    alertmanager:
      image: prom/alertmanager:v0.26.0
      container_name: mindhit-alertmanager
      ports:
        - "9093:9093"
      volumes:
        - ./monitoring/alertmanager/alertmanager.yml:/etc/alertmanager/alertmanager.yml
        - alertmanager_data:/alertmanager
      command:
        - '--config.file=/etc/alertmanager/alertmanager.yml'
        - '--storage.path=/alertmanager'
      networks:
        - mindhit-network
    ```

- [ ] **Alertmanager 설정**
  - [ ] `monitoring/alertmanager/alertmanager.yml`

    ```yaml
    global:
      resolve_timeout: 5m
      slack_api_url: '${SLACK_WEBHOOK_URL}'

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
        slack_configs:
          - channel: '#mindhit-alerts'
            send_resolved: true
            title: '{{ if eq .Status "firing" }}🔥{{ else }}✅{{ end }} {{ .CommonAnnotations.summary }}'
            text: '{{ .CommonAnnotations.description }}'

      - name: 'critical'
        slack_configs:
          - channel: '#mindhit-alerts-critical'
            send_resolved: true
            title: '{{ if eq .Status "firing" }}🚨 CRITICAL{{ else }}✅ RESOLVED{{ end }} {{ .CommonAnnotations.summary }}'
            text: '{{ .CommonAnnotations.description }}'

    inhibit_rules:
      - source_match:
          severity: 'critical'
        target_match:
          severity: 'warning'
        equal: ['alertname']
    ```

- [ ] **환경 변수 파일 업데이트**
  - [ ] `.env.example`에 추가

    ```env
    # Alerting
    SLACK_WEBHOOK_URL=https://hooks.slack.com/services/xxx
    ```

### 검증

```bash
# Alertmanager UI 확인
# http://localhost:9093

# Prometheus Alerts 확인
# http://localhost:9090/alerts

# 테스트 알림 발생 (API 서버 중지)
docker-compose stop api

# 알림 확인 후 재시작
docker-compose start api
```

---

## Phase 12 완료 확인

### 전체 검증 체크리스트

- [ ] **Prometheus 메트릭 수집**
  - [ ] `/metrics` 엔드포인트 응답
  - [ ] HTTP 메트릭 수집 확인
  - [ ] 비즈니스 메트릭 수집 확인

- [ ] **Grafana 대시보드**
  - [ ] 로그인 가능
  - [ ] API Overview 대시보드 표시
  - [ ] Business Metrics 대시보드 표시
  - [ ] Infrastructure 대시보드 표시

- [ ] **로깅 시스템**
  - [ ] 구조화된 JSON 로그 출력
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
| 통합 테스트 | 로그 수집 | Loki API 쿼리 |
| 알림 테스트 | 알림 규칙 | `amtool check-config` |
| 회귀 테스트 | 기존 테스트 통과 | `moon run backend:test` |

```bash
# Phase 12 검증
# 1. 전체 테스트 통과 확인
moon run backend:test

# 2. 모니터링 스택 헬스 체크
curl http://localhost:9090/-/healthy  # Prometheus
curl http://localhost:3001/api/health # Grafana
```

> **Note**: Phase 12는 운영 인프라 설정이므로 기능 테스트보다 시스템 헬스 체크가 중요합니다.

### 산출물 요약

| 항목 | 위치 |
| ---- | ---- |
| 메트릭 미들웨어 | `internal/infrastructure/middleware/metrics.go` |
| 비즈니스 메트릭 | `internal/infrastructure/metrics/business.go` |
| 로거 설정 | `internal/infrastructure/logger/logger.go` |
| Prometheus 설정 | `monitoring/prometheus/prometheus.yml` |
| 알림 규칙 | `monitoring/prometheus/alerts.yml` |
| Grafana 대시보드 | `monitoring/grafana/dashboards/*.json` |
| Alertmanager 설정 | `monitoring/alertmanager/alertmanager.yml` |
| Loki 설정 | `monitoring/loki/loki-config.yaml` |

### 모니터링 스택 요약

| 서비스 | 포트 | 용도 |
|-------|------|------|
| Prometheus | 9090 | 메트릭 수집/저장 |
| Grafana | 3001 | 대시보드/시각화 |
| Alertmanager | 9093 | 알림 관리 |
| Loki | 3100 | 로그 집계 |

### 핵심 메트릭

| 메트릭 | 용도 |
|-------|------|
| `http_requests_total` | API 요청 수 |
| `http_request_duration_seconds` | 응답 시간 |
| `mindhit_active_sessions` | 활성 세션 수 |
| `mindhit_events_received_total` | 이벤트 수신 수 |
| `mindhit_ai_processing_duration_seconds` | AI 처리 시간 |
| `mindhit_db_query_duration_seconds` | DB 쿼리 시간 |

---

## 다음 Phase

Phase 12 완료 후 [Phase 13: 배포/운영](./phase-13-deployment.md)으로 진행하세요.
