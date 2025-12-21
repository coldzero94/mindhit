# Phase 6: 스케줄러 및 백그라운드 작업

## 개요

| 항목 | 내용 |
|-----|------|
| **목표** | gocron 기반 백그라운드 작업 스케줄링 |
| **선행 조건** | Phase 5 완료 |
| **예상 소요** | 2 Steps |
| **결과물** | 주기적 세션 정리 작업 동작 |

---

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 6.1 | gocron 스케줄러 설정 | ⬜ |
| 6.2 | 세션 정리 작업 | ⬜ |
| 6.3 | AI 처리 Job 연동 (Phase 9 이후) | ⬜ |

---

## Step 6.1: gocron 스케줄러 설정

### 체크리스트

- [ ] **의존성 추가**
  ```bash
  go get github.com/go-co-op/gocron/v2
  ```

- [ ] **스케줄러 작성**
  - [ ] `internal/infrastructure/scheduler/scheduler.go`
    ```go
    package scheduler

    import (
        "context"
        "log/slog"
        "time"

        "github.com/go-co-op/gocron/v2"

        "github.com/mindhit/api/ent"
    )

    type Scheduler struct {
        scheduler gocron.Scheduler
        client    *ent.Client
    }

    type Job struct {
        Name     string
        Schedule string // cron expression or duration
        Task     func()
    }

    func New(client *ent.Client) (*Scheduler, error) {
        s, err := gocron.NewScheduler()
        if err != nil {
            return nil, err
        }

        return &Scheduler{
            scheduler: s,
            client:    client,
        }, nil
    }

    func (s *Scheduler) RegisterJob(name string, cronExpr string, task func()) error {
        _, err := s.scheduler.NewJob(
            gocron.CronJob(cronExpr, false),
            gocron.NewTask(func() {
                slog.Info("starting job", "name", name)
                start := time.Now()

                task()

                slog.Info("job completed", "name", name, "duration", time.Since(start))
            }),
            gocron.WithName(name),
        )
        return err
    }

    func (s *Scheduler) RegisterIntervalJob(name string, interval time.Duration, task func()) error {
        _, err := s.scheduler.NewJob(
            gocron.DurationJob(interval),
            gocron.NewTask(func() {
                slog.Info("starting job", "name", name)
                start := time.Now()

                task()

                slog.Info("job completed", "name", name, "duration", time.Since(start))
            }),
            gocron.WithName(name),
        )
        return err
    }

    func (s *Scheduler) Start() {
        s.scheduler.Start()
        slog.Info("scheduler started")
    }

    func (s *Scheduler) Stop() error {
        return s.scheduler.Shutdown()
    }
    ```

- [ ] **main.go에 스케줄러 추가**
  ```go
  import "github.com/mindhit/api/internal/infrastructure/scheduler"

  // Initialize scheduler
  sched, err := scheduler.New(client)
  if err != nil {
      slog.Error("failed to create scheduler", "error", err)
      os.Exit(1)
  }

  // Register jobs (Phase 6.2에서 추가)
  // ...

  // Start scheduler
  sched.Start()
  defer sched.Stop()
  ```

### 검증
```bash
go build ./...
# 컴파일 성공
```

---

## Step 6.2: 세션 정리 작업

### 체크리스트

- [ ] **SessionService에 정리 메서드 추가**
  - [ ] `internal/service/session_service.go`에 추가
    ```go
    // CleanupStaleSessions marks old recording sessions as failed
    func (s *SessionService) CleanupStaleSessions(ctx context.Context, maxAge time.Duration) (int, error) {
        threshold := time.Now().Add(-maxAge)

        // Find stale sessions
        staleSessions, err := s.client.Session.
            Query().
            Where(
                session.StatusIn(session.StatusRecording, session.StatusPaused),
                session.UpdatedAtLT(threshold),
            ).
            All(ctx)

        if err != nil {
            return 0, err
        }

        // Update each to failed status
        for _, sess := range staleSessions {
            _, err := s.client.Session.
                UpdateOne(sess).
                SetStatus(session.StatusFailed).
                Save(ctx)
            if err != nil {
                slog.Error("failed to update stale session", "session_id", sess.ID, "error", err)
            }
        }

        return len(staleSessions), nil
    }

    // GetProcessingSessions returns sessions in processing state
    func (s *SessionService) GetProcessingSessions(ctx context.Context) ([]*ent.Session, error) {
        return s.client.Session.
            Query().
            Where(session.StatusEQ(session.StatusProcessing)).
            All(ctx)
    }

    // CountActiveSessions returns the number of active sessions
    func (s *SessionService) CountActiveSessions(ctx context.Context) (int, error) {
        return s.client.Session.
            Query().
            Where(session.StatusIn(session.StatusRecording, session.StatusPaused)).
            Count(ctx)
    }
    ```

- [ ] **Jobs 패키지 작성**
  - [ ] `internal/jobs/cleanup.go`
    ```go
    package jobs

    import (
        "context"
        "log/slog"
        "time"

        "github.com/mindhit/api/internal/infrastructure/middleware"
        "github.com/mindhit/api/internal/service"
    )

    type CleanupJob struct {
        sessionService *service.SessionService
        maxAge         time.Duration
    }

    func NewCleanupJob(sessionService *service.SessionService, maxAge time.Duration) *CleanupJob {
        return &CleanupJob{
            sessionService: sessionService,
            maxAge:         maxAge,
        }
    }

    func (j *CleanupJob) Run() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
        defer cancel()

        count, err := j.sessionService.CleanupStaleSessions(ctx, j.maxAge)
        if err != nil {
            slog.Error("cleanup job failed", "error", err)
            return
        }

        if count > 0 {
            slog.Info("cleaned up stale sessions", "count", count)
        }
    }
    ```

- [ ] **메트릭 업데이트 Job**
  - [ ] `internal/jobs/metrics.go`
    ```go
    package jobs

    import (
        "context"
        "log/slog"
        "time"

        "github.com/mindhit/api/internal/infrastructure/middleware"
        "github.com/mindhit/api/internal/service"
    )

    type MetricsJob struct {
        sessionService *service.SessionService
    }

    func NewMetricsJob(sessionService *service.SessionService) *MetricsJob {
        return &MetricsJob{sessionService: sessionService}
    }

    func (j *MetricsJob) Run() {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        count, err := j.sessionService.CountActiveSessions(ctx)
        if err != nil {
            slog.Error("metrics job failed", "error", err)
            return
        }

        middleware.SetActiveSessions(count)
    }
    ```

- [ ] **main.go에 jobs 등록**
  ```go
  import "github.com/mindhit/api/internal/jobs"

  // Register cleanup job (hourly)
  cleanupJob := jobs.NewCleanupJob(sessionService, 24*time.Hour)
  if err := sched.RegisterJob("cleanup-stale-sessions", "0 * * * *", cleanupJob.Run); err != nil {
      slog.Error("failed to register cleanup job", "error", err)
  }

  // Register metrics job (every minute)
  metricsJob := jobs.NewMetricsJob(sessionService)
  if err := sched.RegisterIntervalJob("update-metrics", time.Minute, metricsJob.Run); err != nil {
      slog.Error("failed to register metrics job", "error", err)
  }
  ```

### 검증
```bash
# 서버 실행 후 로그 확인
go run ./cmd/server

# 1분 후 메트릭 job 실행 로그 확인
# "starting job" name=update-metrics
# "job completed" name=update-metrics
```

---

## Step 6.3: AI 처리 Job 연동 (Phase 9 이후)

### 목표
Phase 9 완료 후 AI 처리 Job을 스케줄러에 등록

> **Note**: 이 Step은 Phase 9 (AI 마인드맵 생성) 완료 후에 진행합니다.

### 체크리스트

- [ ] **AI 처리 Job 등록** (Phase 9의 `internal/jobs/ai_processing.go` 사용)
  ```go
  import (
      "github.com/mindhit/api/internal/infrastructure/ai"
      "github.com/mindhit/api/internal/infrastructure/crawler"
      "github.com/mindhit/api/internal/jobs"
      "github.com/mindhit/api/internal/service"
  )

  // Initialize AI components (Phase 9에서 구현)
  openaiClient := ai.NewOpenAIClient(cfg.OpenAIAPIKey, cfg.OpenAIModel)
  crawlerClient := crawler.New()

  // Initialize AI services
  summarizeService := service.NewSummarizeService(client, openaiClient, crawlerClient)
  mindmapService := service.NewMindmapService(client, openaiClient)

  // Register AI processing job (every 5 minutes)
  aiProcessingJob := jobs.NewAIProcessingJob(client, summarizeService, mindmapService)
  if err := sched.RegisterIntervalJob("ai-processing", 5*time.Minute, aiProcessingJob.Run); err != nil {
      slog.Error("failed to register ai processing job", "error", err)
  }
  ```

- [ ] **전체 Job 등록 순서 (main.go)**
  ```go
  // 1. Cleanup job (hourly) - 오래된 세션 정리
  cleanupJob := jobs.NewCleanupJob(sessionService, 24*time.Hour)
  sched.RegisterJob("cleanup-stale-sessions", "0 * * * *", cleanupJob.Run)

  // 2. Metrics job (every minute) - 활성 세션 수 업데이트
  metricsJob := jobs.NewMetricsJob(sessionService)
  sched.RegisterIntervalJob("update-metrics", time.Minute, metricsJob.Run)

  // 3. AI processing job (every 5 minutes) - 마인드맵 생성
  aiProcessingJob := jobs.NewAIProcessingJob(client, summarizeService, mindmapService)
  sched.RegisterIntervalJob("ai-processing", 5*time.Minute, aiProcessingJob.Run)
  ```

### 검증
```bash
# 서버 실행
go run ./cmd/server

# 세션 종료 후 5분 이내 로그 확인
# "starting job" name=ai-processing
# "processing session" session_id=...
# "session processing completed" session_id=...
# "job completed" name=ai-processing
```

---

## Job 요약

| Job 이름 | 주기 | 설명 | 의존 Phase |
|---------|------|------|-----------|
| cleanup-stale-sessions | 매시간 (0 * * * *) | 24시간 이상 비활성 세션 정리 | Phase 6 |
| update-metrics | 1분 | 활성 세션 수 Prometheus 메트릭 업데이트 | Phase 5, 6 |
| ai-processing | 5분 | processing 상태 세션의 마인드맵 생성 | Phase 9 |

---

## Phase 6 완료 확인

### 전체 검증 체크리스트

- [ ] 스케줄러 시작 로그
- [ ] 주기적 job 실행 로그
- [ ] `/metrics`에서 `mindhit_sessions_active` 확인

### 산출물 요약

| 항목 | 위치 |
|-----|------|
| 스케줄러 | `internal/infrastructure/scheduler/scheduler.go` |
| Cleanup Job | `internal/jobs/cleanup.go` |
| Metrics Job | `internal/jobs/metrics.go` |

---

## 다음 Phase

Phase 6 완료 후 [Phase 7: Next.js 웹앱](./phase-7-webapp.md)으로 진행하세요.
