# Phase 5: 모니터링 및 인프라

## 개요

| 항목 | 내용 |
|-----|------|
| **목표** | Prometheus 메트릭, 로깅, Makefile 설정 |
| **선행 조건** | Phase 4 완료 |
| **예상 소요** | 3 Steps |
| **결과물** | 모니터링 가능한 서버 인프라 |

---

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 5.1 | Prometheus 메트릭 설정 | ⬜ |
| 5.2 | 로깅 설정 | ⬜ |
| 5.3 | Makefile 작성 | ⬜ |

---

## Step 5.1: Prometheus 메트릭 설정

### 체크리스트

- [ ] **의존성 추가**
  ```bash
  go get github.com/prometheus/client_golang/prometheus
  go get github.com/prometheus/client_golang/prometheus/promhttp
  go get github.com/prometheus/client_golang/prometheus/promauto
  ```

- [ ] **메트릭 미들웨어 작성**
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

    // IncrementEventsProcessed increments the events counter
    func IncrementEventsProcessed(count int) {
        eventsProcessed.Add(float64(count))
    }

    // SetActiveSessions sets the active sessions gauge
    func SetActiveSessions(count int) {
        sessionsActive.Set(float64(count))
    }
    ```

- [ ] **main.go에 메트릭 엔드포인트 추가**
  ```go
  import "github.com/prometheus/client_golang/prometheus/promhttp"

  // Metrics endpoint
  r.GET("/metrics", gin.WrapH(promhttp.Handler()))
  ```

### 검증
```bash
curl http://localhost:8080/metrics | head -20
# mindhit_http_requests_total{...} 등 메트릭 확인
```

---

## Step 5.2: 로깅 설정

### 체크리스트

- [ ] **Logger 패키지 작성**
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
    )

    // Setup initializes the default logger
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

        slog.SetDefault(slog.New(handler))
    }

    // WithContext creates a logger with context values
    func WithContext(ctx context.Context) *slog.Logger {
        logger := slog.Default()

        if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
            logger = logger.With("request_id", requestID)
        }

        if userID, ok := ctx.Value(UserIDKey).(string); ok {
            logger = logger.With("user_id", userID)
        }

        return logger
    }

    // Info logs at info level
    func Info(msg string, args ...any) {
        slog.Info(msg, args...)
    }

    // Error logs at error level
    func Error(msg string, args ...any) {
        slog.Error(msg, args...)
    }

    // Debug logs at debug level
    func Debug(msg string, args ...any) {
        slog.Debug(msg, args...)
    }

    // Warn logs at warn level
    func Warn(msg string, args ...any) {
        slog.Warn(msg, args...)
    }
    ```

- [ ] **Request ID 미들웨어**
  - [ ] `internal/infrastructure/middleware/request_id.go`
    ```go
    package middleware

    import (
        "context"

        "github.com/gin-gonic/gin"
        "github.com/google/uuid"

        "github.com/mindhit/api/internal/infrastructure/logger"
    )

    const RequestIDHeader = "X-Request-ID"

    func RequestID() gin.HandlerFunc {
        return func(c *gin.Context) {
            requestID := c.GetHeader(RequestIDHeader)
            if requestID == "" {
                requestID = uuid.New().String()
            }

            c.Header(RequestIDHeader, requestID)
            ctx := context.WithValue(c.Request.Context(), logger.RequestIDKey, requestID)
            c.Request = c.Request.WithContext(ctx)

            c.Next()
        }
    }
    ```

- [ ] **로깅 미들웨어 업데이트**
  ```go
  func Logging() gin.HandlerFunc {
      return func(c *gin.Context) {
          start := time.Now()
          path := c.Request.URL.Path
          query := c.Request.URL.RawQuery

          c.Next()

          latency := time.Since(start)
          status := c.Writer.Status()

          log := logger.WithContext(c.Request.Context())

          attrs := []any{
              "method", c.Request.Method,
              "path", path,
              "query", query,
              "status", status,
              "latency_ms", latency.Milliseconds(),
              "ip", c.ClientIP(),
              "user_agent", c.Request.UserAgent(),
          }

          if status >= 500 {
              log.Error("request completed", attrs...)
          } else if status >= 400 {
              log.Warn("request completed", attrs...)
          } else {
              log.Info("request completed", attrs...)
          }
      }
  }
  ```

### 검증
```bash
# 개발 환경
ENVIRONMENT=development go run ./cmd/server
# 텍스트 로그 출력

# 프로덕션 환경
ENVIRONMENT=production go run ./cmd/server
# JSON 로그 출력
```

---

## Step 5.3: Makefile 작성

### 체크리스트

- [ ] **Makefile 작성**
  - [ ] `apps/api/Makefile`
    ```makefile
    .PHONY: dev build test lint run generate generate-api migrate-diff migrate-apply migrate-status docker-up docker-down clean help

    # Variables
    BINARY_NAME := server
    BINARY_PATH := ./bin/$(BINARY_NAME)
    OPENAPI_SPEC := ../../packages/protocol/tsp-output/openapi/openapi.yaml

    # Default target
    .DEFAULT_GOAL := help

    ## Development
    dev: ## Run server in development mode
    	@go run ./cmd/server

    run: build ## Build and run server
    	@$(BINARY_PATH)

    ## Build
    build: ## Build the binary
    	@echo "Building..."
    	@go build -o $(BINARY_PATH) ./cmd/server

    build-linux: ## Build for Linux
    	@GOOS=linux GOARCH=amd64 go build -o $(BINARY_PATH)-linux ./cmd/server

    ## Test
    test: ## Run tests
    	@go test -v -race -coverprofile=coverage.out ./...

    test-coverage: test ## Run tests with coverage report
    	@go tool cover -html=coverage.out -o coverage.html
    	@echo "Coverage report: coverage.html"

    ## Lint
    lint: ## Run linter
    	@golangci-lint run

    lint-fix: ## Run linter with auto-fix
    	@golangci-lint run --fix

    ## Code Generation
    generate: ## Generate Ent code
    	@go generate ./ent

    generate-api: ## Generate API code from OpenAPI spec
    	@oapi-codegen -config oapi-codegen.yaml $(OPENAPI_SPEC)

    generate-all: generate generate-api ## Generate all code
    	@echo "All code generated"

    ## Database Migration
    migrate-diff: ## Create migration diff (usage: make migrate-diff name=migration_name)
    	@atlas migrate diff $(name) \
    		--dir "file://ent/migrate/migrations" \
    		--to "ent://ent/schema" \
    		--dev-url "$(DEV_DATABASE_URL)"

    migrate-apply: ## Apply migrations
    	@atlas migrate apply \
    		--dir "file://ent/migrate/migrations" \
    		--url "$(DATABASE_URL)"

    migrate-status: ## Check migration status
    	@atlas migrate status \
    		--dir "file://ent/migrate/migrations" \
    		--url "$(DATABASE_URL)"

    ## Docker
    docker-up: ## Start Docker containers
    	@docker-compose up -d

    docker-down: ## Stop Docker containers
    	@docker-compose down

    docker-logs: ## Show Docker logs
    	@docker-compose logs -f

    ## Cleanup
    clean: ## Clean build artifacts
    	@rm -rf bin/
    	@rm -f coverage.out coverage.html

    ## Help
    help: ## Show this help
    	@echo "Usage: make [target]"
    	@echo ""
    	@echo "Targets:"
    	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
    ```

- [ ] **환경 변수 문서화**
  - [ ] `apps/api/.env.example` 업데이트
    ```
    # Server
    PORT=8080
    ENVIRONMENT=development

    # Database
    DATABASE_URL=postgres://postgres:password@localhost:5432/mindhit?sslmode=disable
    DEV_DATABASE_URL=postgres://postgres:password@localhost:5432/mindhit_dev?sslmode=disable

    # Auth
    JWT_SECRET=your-secret-key-change-in-production

    # Redis
    REDIS_URL=redis://localhost:6379

    # OpenAI (Phase 9)
    OPENAI_API_KEY=sk-...
    ```

### 검증
```bash
cd apps/api
make help
# 사용 가능한 명령어 목록 출력

make build
make test
make lint
```

---

## Phase 5 완료 확인

### 전체 검증 체크리스트

- [ ] `/metrics` 엔드포인트 동작
- [ ] 구조화된 로그 출력
- [ ] Request ID 헤더 추가됨
- [ ] `make help` 명령어 동작

### 산출물 요약

| 항목 | 위치 |
|-----|------|
| 메트릭 미들웨어 | `internal/infrastructure/middleware/metrics.go` |
| 로거 | `internal/infrastructure/logger/logger.go` |
| Request ID | `internal/infrastructure/middleware/request_id.go` |
| Makefile | `apps/api/Makefile` |

---

## 다음 Phase

Phase 5 완료 후 [Phase 6: 스케줄러](./phase-6-scheduler.md)로 진행하세요.
