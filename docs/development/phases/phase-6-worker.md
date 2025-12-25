# Phase 6: Worker 및 Job Queue

## 개요

| 항목 | 내용 |
|-----|------|
| **목표** | Asynq 기반 비동기 작업 처리 Worker 구축 |
| **선행 조건** | Phase 5 완료 |
| **예상 소요** | 3 Steps |
| **결과물** | Worker 서버 및 Job 처리 시스템 동작 |

---

## 아키텍처

```mermaid
flowchart TB
    subgraph API Server
        A[API Handler]
        Q[Asynq Client]
    end

    subgraph Redis
        R[(Job Queue)]
    end

    subgraph Worker Server
        W[Asynq Server]
        H1[MindmapHandler]
        H2[TagExtractionHandler]
        H3[NotificationHandler]
    end

    subgraph External
        AI[OpenAI API]
    end

    A -->|Enqueue Job| Q
    Q --> R
    R --> W
    W --> H1
    W --> H2
    W --> H3
    H1 --> AI
    H2 --> AI
```

---

## 테스트 환경

> **중요**: 모든 테스트는 Docker로 실행 중인 로컬 PostgreSQL과 Redis를 사용합니다.

```bash
# 테스트 실행 전 Docker 서비스 확인
docker ps | grep -E "postgres|redis"

# 테스트 DB URL (기본값)
# postgres://postgres:password@localhost:5432/mindhit_test
# Redis URL: redis://localhost:6379
```

테스트 헬퍼: `internal/testutil/db.go`의 `SetupTestDB(t)` 사용

---

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 6.1 | Asynq 설정 및 Worker 기본 구조 | ⬜ |
| 6.2 | Job 정의 및 Handler 구현 | ⬜ |
| 6.3 | API에서 Job Enqueue 연동 | ⬜ |

---

## Step 6.1: Asynq 설정 및 Worker 기본 구조

### 체크리스트

- [ ] **의존성 추가**

  ```bash
  cd apps/backend
  go get github.com/hibiken/asynq
  ```

- [ ] **Asynq 클라이언트 (Job Enqueue용)**
  - [ ] `internal/infrastructure/queue/client.go`

    ```go
    package queue

    import (
        "github.com/hibiken/asynq"
    )

    type Client struct {
        client *asynq.Client
    }

    func NewClient(redisAddr string) *Client {
        client := asynq.NewClient(asynq.RedisClientOpt{
            Addr: redisAddr,
        })
        return &Client{client: client}
    }

    func (c *Client) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
        return c.client.Enqueue(task, opts...)
    }

    func (c *Client) Close() error {
        return c.client.Close()
    }
    ```

- [ ] **Asynq 서버 설정 (Worker용)**
  - [ ] `internal/infrastructure/queue/server.go`

    ```go
    package queue

    import (
        "log/slog"

        "github.com/hibiken/asynq"
    )

    type Server struct {
        server *asynq.Server
        mux    *asynq.ServeMux
    }

    type ServerConfig struct {
        RedisAddr   string
        Concurrency int
        Queues      map[string]int // queue name -> priority
    }

    func NewServer(cfg ServerConfig) *Server {
        if cfg.Concurrency == 0 {
            cfg.Concurrency = 10
        }
        if cfg.Queues == nil {
            cfg.Queues = map[string]int{
                "critical": 6,
                "default":  3,
                "low":      1,
            }
        }

        server := asynq.NewServer(
            asynq.RedisClientOpt{Addr: cfg.RedisAddr},
            asynq.Config{
                Concurrency: cfg.Concurrency,
                Queues:      cfg.Queues,
                ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
                    slog.Error("task failed",
                        "type", task.Type(),
                        "error", err,
                    )
                }),
            },
        )

        return &Server{
            server: server,
            mux:    asynq.NewServeMux(),
        }
    }

    func (s *Server) HandleFunc(pattern string, handler func(context.Context, *asynq.Task) error) {
        s.mux.HandleFunc(pattern, handler)
    }

    func (s *Server) Run() error {
        slog.Info("starting worker server")
        return s.server.Run(s.mux)
    }

    func (s *Server) Shutdown() {
        s.server.Shutdown()
    }
    ```

- [ ] **Worker 엔트리포인트**
  - [ ] `cmd/worker/main.go`

    ```go
    package main

    import (
        "context"
        "log/slog"
        "os"
        "os/signal"
        "syscall"

        "github.com/mindhit/api/internal/worker/handler"
        "github.com/mindhit/api/internal/infrastructure/config"
        "github.com/mindhit/api/ent"
        "github.com/mindhit/api/internal/infrastructure/queue"

        _ "github.com/lib/pq"
    )

    func main() {
        // Load config
        cfg := config.Load()

        // Setup logger
        slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelInfo,
        })))

        // Connect to database
        client, err := ent.Open("postgres", cfg.DatabaseURL)
        if err != nil {
            slog.Error("failed to connect to database", "error", err)
            os.Exit(1)
        }
        defer client.Close()

        // Run migrations
        if err := client.Schema.Create(context.Background()); err != nil {
            slog.Error("failed to run migrations", "error", err)
            os.Exit(1)
        }

        // Create worker server
        server := queue.NewServer(queue.ServerConfig{
            RedisAddr:   cfg.RedisAddr,
            Concurrency: cfg.WorkerConcurrency,
        })

        // Register handlers
        handler.RegisterHandlers(server, client)

        // Graceful shutdown
        go func() {
            sigCh := make(chan os.Signal, 1)
            signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
            <-sigCh
            slog.Info("shutting down worker")
            server.Shutdown()
        }()

        // Start server
        if err := server.Run(); err != nil {
            slog.Error("worker server error", "error", err)
            os.Exit(1)
        }
    }
    ```

- [ ] **Config 업데이트**
  - [ ] `internal/infrastructure/config/config.go`에 Redis 설정 추가

    ```go
    type Config struct {
        // ... existing fields
        RedisAddr         string
        WorkerConcurrency int
    }

    func Load() *Config {
        return &Config{
            // ... existing fields
            RedisAddr:         getEnv("REDIS_ADDR", "localhost:6379"),
            WorkerConcurrency: getEnvInt("WORKER_CONCURRENCY", 10),
        }
    }
    ```

### 검증

```bash
cd apps/backend
go build ./cmd/worker
# 컴파일 성공
```

---

## Step 6.2: Job 정의 및 Handler 구현

> **에러 처리 가이드**: Worker 에러 처리 패턴은 [09-error-handling.md#7 Worker 에러 처리](../09-error-handling.md#7-worker-에러-처리)를 참조하세요.
>
> - `asynq.SkipRetry`로 재시도 불가 에러 표시
> - 구조화된 `slog` 로깅으로 작업 추적
> - 지수 백오프 재시도 전략

### 체크리스트

- [ ] **Task 타입 정의**
  - [ ] `internal/infrastructure/queue/tasks.go`

    ```go
    package queue

    import (
        "encoding/json"

        "github.com/hibiken/asynq"
    )

    // Task types
    const (
        TypeSessionProcess  = "session:process"
        TypeSessionCleanup  = "session:cleanup"
        TypeURLSummarize    = "url:summarize"
        TypeMindmapGenerate = "mindmap:generate"
    )

    // SessionProcessPayload is the payload for session processing
    type SessionProcessPayload struct {
        SessionID string `json:"session_id"`
    }

    // NewSessionProcessTask creates a new session process task
    func NewSessionProcessTask(sessionID string) (*asynq.Task, error) {
        payload, err := json.Marshal(SessionProcessPayload{SessionID: sessionID})
        if err != nil {
            return nil, err
        }
        return asynq.NewTask(TypeSessionProcess, payload), nil
    }

    // SessionCleanupPayload is the payload for session cleanup
    type SessionCleanupPayload struct {
        MaxAgeHours int `json:"max_age_hours"`
    }

    // NewSessionCleanupTask creates a new session cleanup task
    func NewSessionCleanupTask(maxAgeHours int) (*asynq.Task, error) {
        payload, err := json.Marshal(SessionCleanupPayload{MaxAgeHours: maxAgeHours})
        if err != nil {
            return nil, err
        }
        return asynq.NewTask(TypeSessionCleanup, payload), nil
    }

    // URLSummarizePayload is the payload for URL summarization
    type URLSummarizePayload struct {
        SessionID string `json:"session_id"`
        URL       string `json:"url"`
    }

    // NewURLSummarizeTask creates a new URL summarize task
    func NewURLSummarizeTask(sessionID, url string) (*asynq.Task, error) {
        payload, err := json.Marshal(URLSummarizePayload{
            SessionID: sessionID,
            URL:       url,
        })
        if err != nil {
            return nil, err
        }
        return asynq.NewTask(TypeURLSummarize, payload), nil
    }

    // MindmapGeneratePayload is the payload for mindmap generation
    type MindmapGeneratePayload struct {
        SessionID string `json:"session_id"`
    }

    // NewMindmapGenerateTask creates a new mindmap generate task
    func NewMindmapGenerateTask(sessionID string) (*asynq.Task, error) {
        payload, err := json.Marshal(MindmapGeneratePayload{SessionID: sessionID})
        if err != nil {
            return nil, err
        }
        return asynq.NewTask(TypeMindmapGenerate, payload), nil
    }
    ```

- [ ] **Handler 등록**
  - [ ] `internal/worker/handler/handler.go`

    ```go
    package handler

    import (
        "github.com/mindhit/api/ent"
        "github.com/mindhit/api/internal/infrastructure/queue"
    )

    func RegisterHandlers(server *queue.Server, client *ent.Client) {
        h := &handlers{client: client}

        server.HandleFunc(queue.TypeSessionProcess, h.HandleSessionProcess)
        server.HandleFunc(queue.TypeSessionCleanup, h.HandleSessionCleanup)
        // AI handlers are registered in Phase 9
        // server.HandleFunc(queue.TypeURLSummarize, h.HandleURLSummarize)
        // server.HandleFunc(queue.TypeMindmapGenerate, h.HandleMindmapGenerate)
    }

    type handlers struct {
        client *ent.Client
        // AI services will be added in Phase 9
    }
    ```

- [ ] **Session Process Handler**
  - [ ] `internal/worker/handler/session.go`

    ```go
    package handler

    import (
        "context"
        "encoding/json"
        "fmt"
        "log/slog"

        "github.com/hibiken/asynq"

        "github.com/mindhit/api/ent/session"
        "github.com/mindhit/api/internal/infrastructure/queue"
    )

    func (h *handlers) HandleSessionProcess(ctx context.Context, t *asynq.Task) error {
        var payload queue.SessionProcessPayload
        if err := json.Unmarshal(t.Payload(), &payload); err != nil {
            return fmt.Errorf("failed to unmarshal payload: %w", err)
        }

        slog.Info("processing session", "session_id", payload.SessionID)

        // Get session
        sess, err := h.client.Session.Query().
            Where(session.IDEQ(payload.SessionID)).
            Only(ctx)
        if err != nil {
            return fmt.Errorf("failed to get session: %w", err)
        }

        // Verify session is in processing state
        if sess.Status != session.StatusProcessing {
            slog.Warn("session not in processing state",
                "session_id", payload.SessionID,
                "status", sess.Status,
            )
            return nil // Not an error, just skip
        }

        // TODO: Phase 9에서 AI 처리 로직 추가
        // 1. URL 요약
        // 2. 마인드맵 생성

        slog.Info("session processing completed", "session_id", payload.SessionID)
        return nil
    }
    ```

- [ ] **Session Cleanup Handler**
  - [ ] `internal/worker/handler/cleanup.go`

    ```go
    package handler

    import (
        "context"
        "encoding/json"
        "fmt"
        "log/slog"
        "time"

        "github.com/hibiken/asynq"

        "github.com/mindhit/api/ent/session"
        "github.com/mindhit/api/internal/infrastructure/queue"
    )

    func (h *handlers) HandleSessionCleanup(ctx context.Context, t *asynq.Task) error {
        var payload queue.SessionCleanupPayload
        if err := json.Unmarshal(t.Payload(), &payload); err != nil {
            return fmt.Errorf("failed to unmarshal payload: %w", err)
        }

        slog.Info("starting session cleanup", "max_age_hours", payload.MaxAgeHours)

        threshold := time.Now().Add(-time.Duration(payload.MaxAgeHours) * time.Hour)

        // Find stale sessions
        staleSessions, err := h.client.Session.Query().
            Where(
                session.StatusIn(session.StatusRecording, session.StatusPaused),
                session.UpdatedAtLT(threshold),
            ).
            All(ctx)
        if err != nil {
            return fmt.Errorf("failed to query stale sessions: %w", err)
        }

        // Update each to failed status
        for _, sess := range staleSessions {
            _, err := h.client.Session.UpdateOne(sess).
                SetStatus(session.StatusFailed).
                Save(ctx)
            if err != nil {
                slog.Error("failed to update stale session",
                    "session_id", sess.ID,
                    "error", err,
                )
            }
        }

        slog.Info("session cleanup completed", "cleaned_count", len(staleSessions))
        return nil
    }
    ```

### 검증

```bash
cd apps/backend
go build ./cmd/worker
# 컴파일 성공
```

---

## Step 6.3: API에서 Job Enqueue 연동

### 체크리스트

- [ ] **API에 Queue Client 추가**
  - [ ] `cmd/api/main.go`에 Queue Client 초기화

    ```go
    import "github.com/mindhit/api/internal/infrastructure/queue"

    // Initialize queue client
    queueClient := queue.NewClient(cfg.RedisAddr)
    defer queueClient.Close()

    // Pass to services/controllers that need it
    ```

- [ ] **SessionService에 Enqueue 연동**
  - [ ] `internal/service/session_service.go` 수정

    ```go
    package service

    import (
        "context"
        "log/slog"

        "github.com/hibiken/asynq"

        "github.com/mindhit/api/ent"
        "github.com/mindhit/api/ent/session"
        "github.com/mindhit/api/internal/infrastructure/queue"
    )

    type SessionService struct {
        client      *ent.Client
        queueClient *queue.Client
    }

    func NewSessionService(client *ent.Client, queueClient *queue.Client) *SessionService {
        return &SessionService{
            client:      client,
            queueClient: queueClient,
        }
    }

    // StopSession stops a recording session and enqueues processing
    func (s *SessionService) StopSession(ctx context.Context, sessionID string) (*ent.Session, error) {
        // Get session
        sess, err := s.client.Session.Query().
            Where(session.IDEQ(sessionID)).
            Only(ctx)
        if err != nil {
            return nil, err
        }

        // Update status to processing
        sess, err = s.client.Session.UpdateOne(sess).
            SetStatus(session.StatusProcessing).
            Save(ctx)
        if err != nil {
            return nil, err
        }

        // Enqueue processing task
        task, err := queue.NewSessionProcessTask(sessionID)
        if err != nil {
            slog.Error("failed to create task", "error", err)
            return sess, nil // Don't fail the request
        }

        _, err = s.queueClient.Enqueue(task, asynq.MaxRetry(3))
        if err != nil {
            slog.Error("failed to enqueue task", "error", err)
            return sess, nil // Don't fail the request
        }

        slog.Info("session processing enqueued", "session_id", sessionID)
        return sess, nil
    }
    ```

- [ ] **주기적 Cleanup Task 스케줄링**
  - [ ] `internal/infrastructure/queue/scheduler.go`

    ```go
    package queue

    import (
        "log/slog"

        "github.com/hibiken/asynq"
    )

    type Scheduler struct {
        scheduler *asynq.Scheduler
    }

    func NewScheduler(redisAddr string) (*Scheduler, error) {
        scheduler := asynq.NewScheduler(
            asynq.RedisClientOpt{Addr: redisAddr},
            nil,
        )
        return &Scheduler{scheduler: scheduler}, nil
    }

    func (s *Scheduler) RegisterPeriodicTasks() error {
        // Cleanup stale sessions every hour
        cleanupTask, err := NewSessionCleanupTask(24) // 24 hours
        if err != nil {
            return err
        }

        _, err = s.scheduler.Register("@every 1h", cleanupTask)
        if err != nil {
            return err
        }

        slog.Info("registered periodic cleanup task", "interval", "1h")
        return nil
    }

    func (s *Scheduler) Run() error {
        return s.scheduler.Run()
    }

    func (s *Scheduler) Shutdown() {
        s.scheduler.Shutdown()
    }
    ```

- [ ] **Worker에 Scheduler 통합 (선택적)**
  - [ ] `cmd/worker/main.go`에 스케줄러 추가

    ```go
    // Option 1: Worker와 Scheduler를 함께 실행
    // Create scheduler for periodic tasks
    scheduler, err := queue.NewScheduler(cfg.RedisAddr)
    if err != nil {
        slog.Error("failed to create scheduler", "error", err)
        os.Exit(1)
    }

    if err := scheduler.RegisterPeriodicTasks(); err != nil {
        slog.Error("failed to register periodic tasks", "error", err)
        os.Exit(1)
    }

    // Run scheduler in background
    go func() {
        if err := scheduler.Run(); err != nil {
            slog.Error("scheduler error", "error", err)
        }
    }()
    defer scheduler.Shutdown()
    ```

### 검증

```bash
# go run 모드 (Docker Compose)
moonx infra:dev-up  # Redis 시작

# Worker 실행 (터미널 1)
moonx backend:dev-worker

# API 실행 (터미널 2)
moonx backend:dev-api

# 4. 세션 종료 API 호출 후 Worker 로그 확인
curl -X POST http://localhost:8080/api/v1/sessions/{session_id}/stop

# Worker 로그:
# "processing session" session_id=...
# "session processing completed" session_id=...
```

---

## Asynq 모니터링 (선택)

### Asynqmon 대시보드

```bash
# Asynqmon 설치
go install github.com/hibiken/asynqmon@latest

# 실행
asynqmon --redis-addr=localhost:6379

# 브라우저에서 http://localhost:8080 접속
```

### K8s에서 Asynqmon 배포

```yaml
# deploy/helm/mindhit/templates/asynqmon-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "mindhit.fullname" . }}-asynqmon
spec:
  replicas: 1
  selector:
    matchLabels:
      app: asynqmon
  template:
    metadata:
      labels:
        app: asynqmon
    spec:
      containers:
        - name: asynqmon
          image: hibiken/asynqmon:latest
          ports:
            - containerPort: 8080
          args:
            - "--redis-addr={{ .Values.redis.host }}:{{ .Values.redis.port }}"
```

---

## Phase 6 완료 확인

### 전체 검증 체크리스트

- [ ] Worker 서버 시작 로그
- [ ] API에서 세션 종료 시 Job Enqueue 로그
- [ ] Worker에서 Job 처리 로그
- [ ] 주기적 Cleanup Task 실행 로그

### 테스트 요구사항

| 테스트 유형 | 대상 | 파일 |
| ----------- | ---- | ---- |
| 단위 테스트 | Task 직렬화/역직렬화 | `queue/tasks_test.go` |
| 단위 테스트 | Session Handler 로직 | `handler/session_test.go` |
| 통합 테스트 | Redis 연결 및 Job 처리 | `queue/integration_test.go` |

```bash
# Phase 6 테스트 실행
moonx backend:test -- -run "TestQueue|TestHandler"
```

> **Note**: Worker 테스트는 Redis가 필요합니다. `moonx infra:dev-up`으로 Redis를 먼저 시작하세요.

### 산출물 요약

| 항목 | 위치 |
| ---- | ---- |
| Queue Client | `internal/infrastructure/queue/client.go` |
| Queue Server | `internal/infrastructure/queue/server.go` |
| Task 정의 | `internal/infrastructure/queue/tasks.go` |
| Scheduler | `internal/infrastructure/queue/scheduler.go` |
| Worker 엔트리포인트 | `cmd/worker/main.go` |
| Session Handler | `internal/worker/handler/session.go` |
| Cleanup Handler | `internal/worker/handler/cleanup.go` |
| 테스트 | `internal/infrastructure/queue/*_test.go` |

---

## 다음 Phase

Phase 6 완료 후 [Phase 7: Next.js 웹앱](./phase-7-webapp.md)으로 진행하세요.
