# Phase 10.1: AI 설정 및 로깅

## 개요

| 항목 | 내용 |
|-----|------|
| **목표** | AI 호출 로깅, 동적 설정 관리, Provider Manager 구현 |
| **선행 조건** | Phase 10 완료 (AI Provider 인프라) |
| **예상 소요** | 2 Steps |
| **결과물** | ai_logs 테이블, ai_configs 테이블, Admin API, ProviderManager |

> **Note**: AI Provider 구현체는 [Phase 10](./phase-10-ai.md)에서,
> 마인드맵 생성은 [Phase 10.2](./phase-10.2-mindmap.md)에서 구현합니다.

---

## 설계 원칙

### 하이브리드 설정 구조

```
┌─────────────────────────────────────────────────────────────┐
│                    Hybrid Configuration                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Layer 1: 환경변수 (보안, API 키 전용)                       │
│  ─────────────────────────────────────                      │
│  OPENAI_API_KEY=sk-xxx                                      │
│  CLAUDE_API_KEY=sk-ant-xxx                                  │
│  GEMINI_API_KEY=xxx                                         │
│                                                             │
│  Layer 2: DB ai_configs (동적, 런타임 변경)                  │
│  ──────────────────────────────────────────                 │
│  - 프로바이더/모델 선택                                      │
│  - Fallback 설정                                            │
│  - 온도, 토큰 제한 등 옵션                                   │
│  - 5분 캐시 (성능 최적화)                                    │
│                                                             │
│  → Admin API로 재배포 없이 즉시 변경!                        │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 핵심 기능

1. **AI 호출 로깅**: 모든 AI 호출을 `ai_logs` 테이블에 기록 (디버깅, 비용 추적)
2. **동적 설정**: `ai_configs` 테이블로 재배포 없이 프로바이더/모델 변경
3. **Fallback 지원**: 기본 프로바이더 실패 시 대체 프로바이더 사용
4. **Admin API**: 설정 관리용 API 엔드포인트

---

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 10.1.1 | ai_logs, ai_configs 테이블 및 Service | ⬜ |
| 10.1.2 | Provider Manager 및 Admin API | ⬜ |

---

## Step 10.1.1: ai_logs, ai_configs 테이블 및 Service

### 목표

1. 모든 AI 호출을 `ai_logs` 테이블에 기록하여 디버깅, 비용 추적, 분석에 활용
2. `ai_configs` 테이블을 통해 **재배포 없이** AI 프로바이더/모델 설정 변경

### 체크리스트

- [ ] **AILog Ent 스키마 정의**
  - [ ] `ent/schema/ailog.go`

    ```go
    package schema

    import (
        "time"

        "entgo.io/ent"
        "entgo.io/ent/schema/edge"
        "entgo.io/ent/schema/field"
        "entgo.io/ent/schema/index"
        "github.com/google/uuid"
    )

    // AILog holds the schema definition for the AILog entity.
    type AILog struct {
        ent.Schema
    }

    func (AILog) Fields() []ent.Field {
        return []ent.Field{
            field.UUID("id", uuid.UUID{}).
                Default(uuid.New),

            // 관계
            field.UUID("user_id", uuid.UUID{}).
                Optional().
                Nillable(),
            field.UUID("session_id", uuid.UUID{}).
                Optional().
                Nillable(),

            // 요청 정보
            field.String("task_type").
                NotEmpty().
                Comment("tag_extraction, mindmap, general"),
            field.String("provider").
                NotEmpty().
                Comment("openai, claude, gemini"),
            field.String("model").
                NotEmpty(),
            field.Text("system_prompt").
                Optional(),
            field.Text("user_prompt").
                Optional(),
            field.JSON("messages", []map[string]string{}).
                Optional().
                Comment("Full conversation history"),

            // 응답 정보
            field.Text("thinking").
                Optional().
                Comment("AI reasoning/thinking process"),
            field.Text("content").
                Optional().
                Comment("AI response content (empty on error)"),

            // 토큰 사용량 (정확한 값)
            field.Int("input_tokens").
                Default(0),
            field.Int("output_tokens").
                Default(0),
            field.Int("thinking_tokens").
                Default(0),
            field.Int("total_tokens").
                Default(0),

            // 성능 메트릭
            field.Int64("latency_ms").
                Default(0).
                Comment("Response latency in milliseconds"),
            field.String("request_id").
                Optional().
                Comment("Provider request ID for debugging"),

            // 상태
            field.Enum("status").
                Values("success", "error", "timeout").
                Default("success"),
            field.Text("error_message").
                Optional(),

            // 비용 추적 (cents 단위)
            field.Int("estimated_cost_cents").
                Default(0).
                Comment("Estimated cost in cents"),

            // 추가 메타데이터 (URL, event_id 등 추적용)
            field.JSON("metadata", map[string]interface{}{}).
                Optional().
                Comment("Additional tracking metadata"),

            field.Time("created_at").
                Default(time.Now).
                Immutable(),
        }
    }

    func (AILog) Indexes() []ent.Index {
        return []ent.Index{
            index.Fields("user_id", "created_at"),
            index.Fields("session_id"),
            index.Fields("task_type", "created_at"),
            index.Fields("provider", "model", "created_at"),
            index.Fields("status", "created_at"),
        }
    }

    func (AILog) Edges() []ent.Edge {
        return []ent.Edge{
            edge.From("user", User.Type).
                Ref("ai_logs").
                Field("user_id").
                Unique(),
            edge.From("session", Session.Type).
                Ref("ai_logs").
                Field("session_id").
                Unique(),
        }
    }
    ```

- [ ] **AILogService 구현**
  - [ ] `internal/service/ailog_service.go`

    ```go
    package service

    import (
        "context"

        "github.com/google/uuid"
        "github.com/mindhit/api/ent"
        "github.com/mindhit/api/ent/ailog"
        "github.com/mindhit/api/internal/infrastructure/ai"
    )

    // AILogService manages AI request logging.
    type AILogService struct {
        client *ent.Client
    }

    // NewAILogService creates a new AILogService.
    func NewAILogService(client *ent.Client) *AILogService {
        return &AILogService{client: client}
    }

    // AILogRequest represents the data needed to create an AI log entry.
    type AILogRequest struct {
        UserID       *uuid.UUID
        SessionID    *uuid.UUID
        TaskType     ai.TaskType
        Request      ai.ChatRequest
        Response     *ai.ChatResponse
        ErrorMessage string
    }

    // Log creates an AI log entry from request and response.
    func (s *AILogService) Log(ctx context.Context, req AILogRequest) (*ent.AILog, error) {
        status := "success"
        if req.ErrorMessage != "" {
            status = "error"
        }

        builder := s.client.AILog.Create().
            SetTaskType(string(req.TaskType)).
            SetProvider(string(req.Response.Provider)).
            SetModel(req.Response.Model).
            SetContent(req.Response.Content).
            SetInputTokens(req.Response.InputTokens).
            SetOutputTokens(req.Response.OutputTokens).
            SetThinkingTokens(req.Response.ThinkingTokens).
            SetTotalTokens(req.Response.TotalTokens).
            SetLatencyMs(req.Response.LatencyMs).
            SetStatus(ailog.Status(status))

        if req.UserID != nil {
            builder.SetUserID(*req.UserID)
        }
        if req.SessionID != nil {
            builder.SetSessionID(*req.SessionID)
        }
        if req.Request.SystemPrompt != "" {
            builder.SetSystemPrompt(req.Request.SystemPrompt)
        }
        if req.Request.UserPrompt != "" {
            builder.SetUserPrompt(req.Request.UserPrompt)
        }
        if req.Response.Thinking != "" {
            builder.SetThinking(req.Response.Thinking)
        }
        if req.Response.RequestID != "" {
            builder.SetRequestID(req.Response.RequestID)
        }
        if req.ErrorMessage != "" {
            builder.SetErrorMessage(req.ErrorMessage)
        }

        // Calculate estimated cost
        cost := s.estimateCost(req.Response)
        builder.SetEstimatedCostCents(cost)

        return builder.Save(ctx)
    }

    // estimateCost calculates cost in cents based on provider and tokens.
    func (s *AILogService) estimateCost(resp *ai.ChatResponse) int {
        // Pricing per 1M tokens (approximate, in cents)
        pricing := map[ai.ProviderType]struct{ input, output int }{
            ai.ProviderOpenAI: {input: 250, output: 1000},  // GPT-4o
            ai.ProviderClaude: {input: 300, output: 1500},  // Claude 3.5 Sonnet
            ai.ProviderGemini: {input: 35, output: 105},    // Gemini 1.5 Flash
        }

        p, ok := pricing[resp.Provider]
        if !ok {
            return 0
        }

        inputCost := (resp.InputTokens * p.input) / 1000000
        outputCost := (resp.OutputTokens * p.output) / 1000000

        return inputCost + outputCost
    }

    // GetBySession retrieves all AI logs for a session.
    func (s *AILogService) GetBySession(ctx context.Context, sessionID uuid.UUID) ([]*ent.AILog, error) {
        return s.client.AILog.Query().
            Where(ailog.SessionIDEQ(sessionID)).
            Order(ent.Asc(ailog.FieldCreatedAt)).
            All(ctx)
    }

    // UsageStats holds aggregated usage statistics.
    type UsageStats struct {
        TotalTokens  int `json:"total_tokens"`
        TotalCost    int `json:"total_cost_cents"`
        RequestCount int `json:"request_count"`
    }

    // GetUsageStats returns token usage statistics for a user.
    func (s *AILogService) GetUsageStats(ctx context.Context, userID uuid.UUID) (*UsageStats, error) {
        var stats UsageStats
        err := s.client.AILog.Query().
            Where(ailog.UserIDEQ(userID)).
            Aggregate(
                ent.Sum(ailog.FieldTotalTokens),
                ent.Sum(ailog.FieldEstimatedCostCents),
                ent.Count(),
            ).
            Scan(ctx, &stats)
        return &stats, err
    }
    ```

- [ ] **AIConfig Ent 스키마 정의**
  - [ ] `ent/schema/aiconfig.go`

    ```go
    package schema

    import (
        "time"

        "entgo.io/ent"
        "entgo.io/ent/schema/field"
        "entgo.io/ent/schema/index"
    )

    // AIConfig holds dynamic AI provider configuration.
    // API 키는 환경변수에서 관리, 프로바이더/모델 선택만 DB에서 관리.
    type AIConfig struct {
        ent.Schema
    }

    func (AIConfig) Fields() []ent.Field {
        return []ent.Field{
            // Task 유형 (unique key)
            field.String("task_type").
                NotEmpty().
                Unique().
                Comment("Task type: 'default', 'tag_extraction', 'mindmap'"),

            // 프로바이더 설정
            field.String("provider").
                NotEmpty().
                Comment("AI provider: 'openai', 'claude', 'gemini'"),
            field.String("model").
                NotEmpty().
                Comment("Model name: 'gpt-4o', 'claude-sonnet-4', 'gemini-2.0-flash'"),

            // Fallback 프로바이더 목록
            field.JSON("fallback_providers", []string{}).
                Optional().
                Comment("Ordered list of fallback providers"),

            // 옵션
            field.Float("temperature").
                Default(0.7).
                Comment("Model temperature (0.0-2.0)"),
            field.Int("max_tokens").
                Default(4096).
                Comment("Max output tokens"),
            field.Int("thinking_budget").
                Default(0).
                Comment("Extended thinking token budget (Claude)"),
            field.Bool("json_mode").
                Default(false).
                Comment("Force JSON output"),

            // 활성화 상태
            field.Bool("enabled").
                Default(true).
                Comment("Whether this config is active"),

            // 감사 필드
            field.String("updated_by").
                Optional().
                Comment("Admin who last updated this config"),
            field.Time("created_at").
                Default(time.Now).
                Immutable(),
            field.Time("updated_at").
                Default(time.Now).
                UpdateDefault(time.Now),
        }
    }

    func (AIConfig) Indexes() []ent.Index {
        return []ent.Index{
            index.Fields("task_type").Unique(),
            index.Fields("provider", "model"),
        }
    }
    ```

- [ ] **AIConfigService 구현** (DB CRUD + 캐싱)
  - [ ] `internal/service/aiconfig_service.go`

    ```go
    package service

    import (
        "context"
        "sync"
        "time"

        "github.com/mindhit/api/ent"
        "github.com/mindhit/api/ent/aiconfig"
        "github.com/mindhit/api/internal/infrastructure/ai"
    )

    // AIConfigService manages AI provider configuration in DB with caching.
    type AIConfigService struct {
        client *ent.Client

        // In-memory cache
        cache     map[string]*ent.AIConfig
        cacheMu   sync.RWMutex
        cacheTime time.Time
        cacheTTL  time.Duration
    }

    // NewAIConfigService creates a new AIConfigService.
    func NewAIConfigService(client *ent.Client) *AIConfigService {
        return &AIConfigService{
            client:   client,
            cache:    make(map[string]*ent.AIConfig),
            cacheTTL: 5 * time.Minute,
        }
    }

    // GetConfigForTask returns the config for a specific task type.
    func (s *AIConfigService) GetConfigForTask(ctx context.Context, taskType ai.TaskType) (*ent.AIConfig, error) {
        s.cacheMu.RLock()
        if time.Since(s.cacheTime) < s.cacheTTL {
            if cfg, ok := s.cache[string(taskType)]; ok {
                s.cacheMu.RUnlock()
                return cfg, nil
            }
        }
        s.cacheMu.RUnlock()

        return s.refreshCache(ctx, string(taskType))
    }

    // refreshCache loads config from DB and updates cache.
    func (s *AIConfigService) refreshCache(ctx context.Context, taskType string) (*ent.AIConfig, error) {
        cfg, err := s.client.AIConfig.Query().
            Where(aiconfig.TaskTypeEQ(taskType)).
            Where(aiconfig.EnabledEQ(true)).
            Only(ctx)

        if err != nil {
            if ent.IsNotFound(err) && taskType != "default" {
                return s.refreshCache(ctx, "default")
            }
            return nil, err
        }

        s.cacheMu.Lock()
        s.cache[taskType] = cfg
        s.cacheTime = time.Now()
        s.cacheMu.Unlock()

        return cfg, nil
    }

    // InvalidateCache clears the cache.
    func (s *AIConfigService) InvalidateCache() {
        s.cacheMu.Lock()
        s.cache = make(map[string]*ent.AIConfig)
        s.cacheTime = time.Time{}
        s.cacheMu.Unlock()
    }

    // GetAll returns all AI configs.
    func (s *AIConfigService) GetAll(ctx context.Context) ([]*ent.AIConfig, error) {
        return s.client.AIConfig.Query().
            Order(ent.Asc(aiconfig.FieldTaskType)).
            All(ctx)
    }

    // UpsertAIConfigRequest is the request for creating/updating AI config.
    type UpsertAIConfigRequest struct {
        TaskType          string   `json:"task_type"`
        Provider          string   `json:"provider"`
        Model             string   `json:"model"`
        FallbackProviders []string `json:"fallback_providers,omitempty"`
        Temperature       float64  `json:"temperature"`
        MaxTokens         int      `json:"max_tokens"`
        ThinkingBudget    int      `json:"thinking_budget,omitempty"`
        JSONMode          bool     `json:"json_mode"`
        Enabled           bool     `json:"enabled"`
        UpdatedBy         string   `json:"updated_by,omitempty"`
    }

    // Upsert creates or updates an AI config.
    func (s *AIConfigService) Upsert(ctx context.Context, req UpsertAIConfigRequest) (*ent.AIConfig, error) {
        id, err := s.client.AIConfig.Create().
            SetTaskType(req.TaskType).
            SetProvider(req.Provider).
            SetModel(req.Model).
            SetFallbackProviders(req.FallbackProviders).
            SetTemperature(req.Temperature).
            SetMaxTokens(req.MaxTokens).
            SetThinkingBudget(req.ThinkingBudget).
            SetJSONMode(req.JSONMode).
            SetEnabled(req.Enabled).
            SetUpdatedBy(req.UpdatedBy).
            OnConflictColumns(aiconfig.FieldTaskType).
            UpdateNewValues().
            ID(ctx)

        if err != nil {
            return nil, err
        }

        s.InvalidateCache()
        return s.client.AIConfig.Get(ctx, id)
    }

    // Delete removes an AI config.
    func (s *AIConfigService) Delete(ctx context.Context, taskType string) error {
        _, err := s.client.AIConfig.Delete().
            Where(aiconfig.TaskTypeEQ(taskType)).
            Exec(ctx)

        if err == nil {
            s.InvalidateCache()
        }
        return err
    }

    // SeedDefaultConfigs creates default AI configs if they don't exist.
    func (s *AIConfigService) SeedDefaultConfigs(ctx context.Context) error {
        defaults := []UpsertAIConfigRequest{
            {
                TaskType:          "default",
                Provider:          "openai",
                Model:             "gpt-4o",
                FallbackProviders: []string{"gemini", "claude"},
                Temperature:       0.7,
                MaxTokens:         4096,
                Enabled:           true,
            },
            {
                TaskType:          "tag_extraction",
                Provider:          "gemini",
                Model:             "gemini-2.0-flash",
                FallbackProviders: []string{"openai"},
                Temperature:       0.3,
                MaxTokens:         1024,
                JSONMode:          true,
                Enabled:           true,
            },
            {
                TaskType:          "mindmap",
                Provider:          "claude",
                Model:             "claude-sonnet-4-20250514",
                FallbackProviders: []string{"openai"},
                Temperature:       0.5,
                MaxTokens:         8192,
                ThinkingBudget:    10000,
                JSONMode:          true,
                Enabled:           true,
            },
        }

        for _, cfg := range defaults {
            exists, _ := s.client.AIConfig.Query().
                Where(aiconfig.TaskTypeEQ(cfg.TaskType)).
                Exist(ctx)
            if !exists {
                if _, err := s.Upsert(ctx, cfg); err != nil {
                    return err
                }
            }
        }

        return nil
    }
    ```

### 검증

```bash
cd apps/backend
moonx backend:generate   # Ent 코드 생성
moonx backend:migrate-diff  # 마이그레이션 생성
go build ./...
```

---

## Step 10.1.2: Provider Manager 및 Admin API

### 목표

1. 여러 Provider를 관리하고 Task 유형에 따라 적절한 Provider 선택
2. AI 호출 결과를 자동으로 로깅
3. Admin API로 설정 관리

### 체크리스트

- [ ] **환경 변수 설정** (API 키만)

  ```env
  # AI Provider API Keys (보안: 환경변수로만 관리)
  OPENAI_API_KEY=sk-...
  GEMINI_API_KEY=...
  CLAUDE_API_KEY=sk-ant-...
  ```

- [ ] **Config 업데이트**
  - [ ] `internal/infrastructure/config/config.go`에 AI 설정 추가

    ```go
    type Config struct {
        // ... 기존 필드
        AI AIConfig
    }

    // AIConfig holds API keys only. Provider/model selection is in DB.
    type AIConfig struct {
        OpenAIAPIKey  string `json:"openai_api_key"`
        GeminiAPIKey  string `json:"gemini_api_key"`
        ClaudeAPIKey  string `json:"claude_api_key"`
    }

    func LoadAIConfig() AIConfig {
        return AIConfig{
            OpenAIAPIKey:  os.Getenv("OPENAI_API_KEY"),
            GeminiAPIKey:  os.Getenv("GEMINI_API_KEY"),
            ClaudeAPIKey:  os.Getenv("CLAUDE_API_KEY"),
        }
    }
    ```

- [ ] **Provider Manager 구현**
  - [ ] `internal/infrastructure/ai/manager.go`

    ```go
    package ai

    import (
        "context"
        "fmt"
        "log/slog"
        "sync"

        "github.com/google/uuid"
        "github.com/mindhit/api/ent"
        "github.com/mindhit/api/internal/service"
    )

    // ProviderManager manages AI providers with DB-based config and auto-logging.
    type ProviderManager struct {
        providers     map[ProviderType]AIProvider
        configService *service.AIConfigService
        logService    *service.AILogService
        apiKeys       map[ProviderType]string
        mu            sync.RWMutex
    }

    // NewProviderManager creates a ProviderManager with environment API keys.
    func NewProviderManager(
        ctx context.Context,
        cfg AIConfig,
        configService *service.AIConfigService,
        logService *service.AILogService,
    ) (*ProviderManager, error) {
        pm := &ProviderManager{
            providers:     make(map[ProviderType]AIProvider),
            configService: configService,
            logService:    logService,
            apiKeys: map[ProviderType]string{
                ProviderOpenAI: cfg.OpenAIAPIKey,
                ProviderGemini: cfg.GeminiAPIKey,
                ProviderClaude: cfg.ClaudeAPIKey,
            },
        }

        // Initialize all providers with API keys
        if cfg.OpenAIAPIKey != "" {
            pm.providers[ProviderOpenAI] = NewOpenAIProvider(ProviderConfig{
                Type:   ProviderOpenAI,
                APIKey: cfg.OpenAIAPIKey,
                Model:  "gpt-4o",
            })
            slog.Info("initialized ai provider", "provider", "openai")
        }

        if cfg.GeminiAPIKey != "" {
            gemini, err := NewGeminiProvider(ctx, ProviderConfig{
                Type:   ProviderGemini,
                APIKey: cfg.GeminiAPIKey,
                Model:  "gemini-2.0-flash",
            })
            if err != nil {
                slog.Warn("failed to initialize gemini provider", "error", err)
            } else {
                pm.providers[ProviderGemini] = gemini
                slog.Info("initialized ai provider", "provider", "gemini")
            }
        }

        if cfg.ClaudeAPIKey != "" {
            pm.providers[ProviderClaude] = NewClaudeProvider(ProviderConfig{
                Type:   ProviderClaude,
                APIKey: cfg.ClaudeAPIKey,
                Model:  "claude-sonnet-4-20250514",
            })
            slog.Info("initialized ai provider", "provider", "claude")
        }

        if len(pm.providers) == 0 {
            return nil, fmt.Errorf("no ai providers configured (missing API keys)")
        }

        slog.Info("provider manager initialized", "available_providers", len(pm.providers))
        return pm, nil
    }

    // Chat executes a request using DB-configured provider for the task.
    func (pm *ProviderManager) Chat(ctx context.Context, task TaskType, req ChatRequest) (*ChatResponse, error) {
        cfg, err := pm.configService.GetConfigForTask(ctx, task)
        if err != nil {
            return nil, fmt.Errorf("failed to get config for task %s: %w", task, err)
        }

        // Apply DB config to request options
        req.Options.Temperature = cfg.Temperature
        req.Options.MaxTokens = cfg.MaxTokens
        req.Options.JSONMode = cfg.JSONMode
        if cfg.ThinkingBudget > 0 {
            req.Options.EnableThinking = true
            req.Options.ThinkingBudget = cfg.ThinkingBudget
        }

        providers := pm.getProvidersFromConfig(cfg)
        if len(providers) == 0 {
            return nil, fmt.Errorf("no available providers for task %s", task)
        }

        var lastErr error
        for _, provider := range providers {
            slog.Debug("attempting ai request",
                "provider", provider.Type(),
                "model", cfg.Model,
                "task", task,
            )

            resp, err := provider.Chat(ctx, req)
            if err == nil {
                pm.logRequest(ctx, task, req, resp, "")
                slog.Info("ai request successful",
                    "provider", resp.Provider,
                    "model", resp.Model,
                    "tokens", resp.TotalTokens,
                    "latency_ms", resp.LatencyMs,
                )
                return resp, nil
            }

            pm.logRequest(ctx, task, req, nil, err.Error())
            lastErr = err
            slog.Warn("ai provider failed, trying fallback",
                "provider", provider.Type(),
                "error", err,
            )
        }

        return nil, fmt.Errorf("all ai providers failed, last error: %w", lastErr)
    }

    // getProvidersFromConfig returns providers based on DB config.
    func (pm *ProviderManager) getProvidersFromConfig(cfg *ent.AIConfig) []AIProvider {
        pm.mu.RLock()
        defer pm.mu.RUnlock()

        var result []AIProvider
        seen := make(map[ProviderType]bool)

        primary := ProviderType(cfg.Provider)
        if p, ok := pm.providers[primary]; ok {
            result = append(result, p)
            seen[primary] = true
        }

        for _, fb := range cfg.FallbackProviders {
            pt := ProviderType(fb)
            if !seen[pt] {
                if p, ok := pm.providers[pt]; ok {
                    result = append(result, p)
                    seen[pt] = true
                }
            }
        }

        return result
    }

    // logRequest logs the AI request to ai_logs table.
    func (pm *ProviderManager) logRequest(
        ctx context.Context,
        task TaskType,
        req ChatRequest,
        resp *ChatResponse,
        errMsg string,
    ) {
        if pm.logService == nil {
            return
        }

        logReq := service.AILogRequest{
            TaskType:     task,
            Request:      req,
            Response:     resp,
            ErrorMessage: errMsg,
        }

        if userID, ok := req.Metadata["user_id"]; ok {
            if uid, err := uuid.Parse(userID); err == nil {
                logReq.UserID = &uid
            }
        }
        if sessionID, ok := req.Metadata["session_id"]; ok {
            if sid, err := uuid.Parse(sessionID); err == nil {
                logReq.SessionID = &sid
            }
        }

        if _, err := pm.logService.Log(ctx, logReq); err != nil {
            slog.Error("failed to log ai request", "error", err)
        }
    }

    // Close closes all providers.
    func (pm *ProviderManager) Close() error {
        pm.mu.Lock()
        defer pm.mu.Unlock()

        for _, provider := range pm.providers {
            if err := provider.Close(); err != nil {
                slog.Warn("failed to close provider", "provider", provider.Type(), "error", err)
            }
        }
        return nil
    }

    // GetAvailableProviders returns list of configured providers.
    func (pm *ProviderManager) GetAvailableProviders() []ProviderType {
        pm.mu.RLock()
        defer pm.mu.RUnlock()

        var result []ProviderType
        for pt := range pm.providers {
            result = append(result, pt)
        }
        return result
    }
    ```

- [ ] **Admin AI Controller 구현**
  - [ ] `internal/controller/admin_ai_controller.go`

    ```go
    package controller

    import (
        "net/http"

        "github.com/gin-gonic/gin"
        "github.com/mindhit/api/internal/generated"
        "github.com/mindhit/api/internal/service"
    )

    // AdminAIController handles AI configuration management.
    type AdminAIController struct {
        configService *service.AIConfigService
    }

    // NewAdminAIController creates a new AdminAIController.
    func NewAdminAIController(configService *service.AIConfigService) *AdminAIController {
        return &AdminAIController{configService: configService}
    }

    // GetAIConfigs returns all AI configurations.
    func (c *AdminAIController) GetAIConfigs(ctx *gin.Context) {
        configs, err := c.configService.GetAll(ctx)
        if err != nil {
            ctx.JSON(http.StatusInternalServerError, generated.Error{
                Code:    "INTERNAL_ERROR",
                Message: "Failed to get AI configs",
            })
            return
        }

        response := make([]generated.AIConfigResponse, len(configs))
        for i, cfg := range configs {
            response[i] = generated.AIConfigResponse{
                TaskType:          cfg.TaskType,
                Provider:          cfg.Provider,
                Model:             cfg.Model,
                FallbackProviders: cfg.FallbackProviders,
                Temperature:       cfg.Temperature,
                MaxTokens:         cfg.MaxTokens,
                ThinkingBudget:    cfg.ThinkingBudget,
                JsonMode:          cfg.JSONMode,
                Enabled:           cfg.Enabled,
                UpdatedBy:         cfg.UpdatedBy,
                UpdatedAt:         cfg.UpdatedAt,
            }
        }

        ctx.JSON(http.StatusOK, response)
    }

    // UpsertAIConfig creates or updates an AI configuration.
    func (c *AdminAIController) UpsertAIConfig(ctx *gin.Context) {
        taskType := ctx.Param("task_type")

        var req generated.AIConfigUpdateRequest
        if err := ctx.ShouldBindJSON(&req); err != nil {
            ctx.JSON(http.StatusBadRequest, generated.Error{
                Code:    "INVALID_REQUEST",
                Message: err.Error(),
            })
            return
        }

        adminEmail := getUserEmail(ctx)

        cfg, err := c.configService.Upsert(ctx, service.UpsertAIConfigRequest{
            TaskType:          taskType,
            Provider:          req.Provider,
            Model:             req.Model,
            FallbackProviders: req.FallbackProviders,
            Temperature:       req.Temperature,
            MaxTokens:         req.MaxTokens,
            ThinkingBudget:    req.ThinkingBudget,
            JSONMode:          req.JsonMode,
            Enabled:           req.Enabled,
            UpdatedBy:         adminEmail,
        })

        if err != nil {
            ctx.JSON(http.StatusInternalServerError, generated.Error{
                Code:    "INTERNAL_ERROR",
                Message: "Failed to update AI config",
            })
            return
        }

        ctx.JSON(http.StatusOK, generated.AIConfigResponse{
            TaskType:          cfg.TaskType,
            Provider:          cfg.Provider,
            Model:             cfg.Model,
            FallbackProviders: cfg.FallbackProviders,
            Temperature:       cfg.Temperature,
            MaxTokens:         cfg.MaxTokens,
            ThinkingBudget:    cfg.ThinkingBudget,
            JsonMode:          cfg.JSONMode,
            Enabled:           cfg.Enabled,
            UpdatedBy:         cfg.UpdatedBy,
            UpdatedAt:         cfg.UpdatedAt,
        })
    }

    // DeleteAIConfig deletes an AI configuration.
    func (c *AdminAIController) DeleteAIConfig(ctx *gin.Context) {
        taskType := ctx.Param("task_type")

        if taskType == "default" {
            ctx.JSON(http.StatusBadRequest, generated.Error{
                Code:    "CANNOT_DELETE_DEFAULT",
                Message: "Cannot delete default configuration",
            })
            return
        }

        if err := c.configService.Delete(ctx, taskType); err != nil {
            ctx.JSON(http.StatusInternalServerError, generated.Error{
                Code:    "INTERNAL_ERROR",
                Message: "Failed to delete AI config",
            })
            return
        }

        ctx.Status(http.StatusNoContent)
    }
    ```

- [ ] **TypeSpec Admin API 정의**
  - [ ] `packages/protocol/src/admin/ai.tsp`

    ```typespec
    import "@typespec/http";
    import "../common/models.tsp";

    using TypeSpec.Http;

    @route("/admin/ai")
    @tag("Admin - AI")
    namespace Admin.AI;

    model AIConfigResponse {
        task_type: string;
        provider: string;
        model: string;
        fallback_providers?: string[];
        temperature: float64;
        max_tokens: int32;
        thinking_budget?: int32;
        json_mode: boolean;
        enabled: boolean;
        updated_by?: string;
        updated_at: utcDateTime;
    }

    model AIConfigUpdateRequest {
        provider: string;
        model: string;
        fallback_providers?: string[];
        temperature?: float64 = 0.7;
        max_tokens?: int32 = 4096;
        thinking_budget?: int32;
        json_mode?: boolean = false;
        enabled?: boolean = true;
    }

    @route("/configs")
    interface AIConfigs {
        @get list(): AIConfigResponse[];
        @put @route("/{task_type}") upsert(@path task_type: string, @body config: AIConfigUpdateRequest): AIConfigResponse;
        @delete @route("/{task_type}") delete(@path task_type: string): void;
    }
    ```

- [ ] **라우터 등록**
  - [ ] `cmd/api/main.go`

    ```go
    // Admin routes (require admin role)
    adminGroup := router.Group("/v1/admin")
    adminGroup.Use(authMiddleware, adminMiddleware)
    {
        adminAI := adminGroup.Group("/ai")
        {
            adminAI.GET("/configs", adminAIController.GetAIConfigs)
            adminAI.PUT("/configs/:task_type", adminAIController.UpsertAIConfig)
            adminAI.DELETE("/configs/:task_type", adminAIController.DeleteAIConfig)
        }
    }
    ```

### 사용 예시

```bash
# 1. 현재 AI 설정 조회
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8080/v1/admin/ai/configs

# 2. 태그 추출 프로바이더를 Gemini에서 OpenAI로 변경
curl -X PUT -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "openai",
    "model": "gpt-4o-mini",
    "fallback_providers": ["gemini"],
    "temperature": 0.3,
    "max_tokens": 1024,
    "json_mode": true,
    "enabled": true
  }' \
  http://localhost:8080/v1/admin/ai/configs/tag_extraction

# → 재배포 없이 즉시 적용! (5분 캐시 후 갱신)
```

### 검증

```bash
cd apps/backend
go build ./...
pnpm run generate
go test ./internal/service/aiconfig_service_test.go
```

---

## 다음 단계

Phase 10.1 완료 후 [Phase 10.2: 마인드맵 생성](./phase-10.2-mindmap.md)으로 진행:

- 태그 추출 Worker Handler
- 마인드맵 생성 Worker Handler
- UsageService 연동 (토큰 측정)
