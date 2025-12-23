# Phase 10: AI 마인드맵 생성

## 개요

| 항목 | 내용 |
|-----|------|
| **목표** | 다중 AI 프로바이더(OpenAI, Google Gemini, Anthropic Claude)를 지원하는 태그 추출 및 마인드맵 생성 |
| **선행 조건** | Phase 9 완료 (플랜 및 사용량 시스템) |
| **예상 소요** | 5 Steps |
| **결과물** | 페이지 방문 시 태그 추출, 세션 종료 시 관계도 JSON 생성 |

> **Note**: 이 Phase에서는 Phase 9의 UsageService를 연동하여 AI 호출 시 토큰 사용량을 추적합니다.

---

## 마인드맵 생성 알고리즘

### 핵심 원칙

1. **페이지 방문 시**: LLM으로 태그/키워드 추출 (페이지당 1회, 중복 URL은 재사용)
2. **세션 종료 시**: 추출된 태그들을 기반으로 LLM이 관계도 JSON 생성 (세션당 1회)

### 처리 흐름 (Asynq Worker 기반)

```mermaid
sequenceDiagram
    participant EXT as Extension
    participant API as API Server
    participant Redis as Redis Queue
    participant Worker as Asynq Worker
    participant DB as Database
    participant AI as AI Provider

    Note over EXT,AI: 1. 페이지 방문 시 (비동기 처리)
    EXT->>API: 이벤트 배치 전송 (URL, content)
    API->>DB: URL 중복 체크 (url_hash)
    alt URL이 새로운 경우
        API->>Redis: Enqueue TagExtraction Task
        Redis->>Worker: Consume Task
        Worker->>AI: 태그 추출 요청 (content)
        AI-->>Worker: { tags: [...], summary: "..." }
        Worker->>DB: urls 테이블에 tags, summary 저장
    else URL이 이미 존재
        API->>DB: 기존 tags 재사용
    end

    Note over EXT,AI: 2. 세션 종료 시 (비동기 처리)
    EXT->>API: 세션 Stop 요청
    API->>DB: 세션 상태 → processing
    API->>Redis: Enqueue MindmapGenerate Task
    Redis->>Worker: Consume Task
    Worker->>DB: 세션의 모든 URL + tags 조회
    Worker->>AI: 관계도 생성 요청 (tags 목록)
    AI-->>Worker: { nodes: [...], edges: [...] }
    Worker->>DB: mindmap_graphs 저장
    Worker->>DB: 세션 상태 → completed
```

### 태그 추출 (페이지당)

| 항목 | 설명 |
|-----|------|
| **트리거** | 이벤트 배치 수신 시 새로운 URL 감지 → Asynq Task Enqueue |
| **입력** | 페이지 제목, 콘텐츠 (최대 10,000자) |
| **출력** | 3-5개 태그, 1-2문장 요약 |
| **저장** | `urls.tags`, `urls.summary` |
| **중복 처리** | url_hash로 중복 체크, 기존 URL은 재처리 안 함 |

**프롬프트 예시:**

```
웹 페이지를 분석하고 다음을 추출하세요:
1. 핵심 태그 3-5개 (한국어, 명사형)
2. 1-2문장 요약

JSON 형식으로 응답:
{
  "tags": ["태그1", "태그2", "태그3"],
  "summary": "페이지 요약"
}
```

### 관계도 생성 (세션당)

| 항목 | 설명 |
|-----|------|
| **트리거** | 세션 종료 (Stop) 시 → Asynq Task Enqueue |
| **입력** | 세션의 모든 URL + tags + 체류시간 + 하이라이트 |
| **출력** | 마인드맵 JSON (nodes, edges) |
| **저장** | `mindmap_graphs` 테이블 |

### 비용 최적화

| 전략 | 설명 |
|-----|------|
| **URL 중복 제거** | 같은 URL은 태그 1번만 추출 (url_hash 기반) |
| **비동기 처리** | Asynq Worker에서 처리하여 API 응답 속도 유지 |
| **경량 모델 사용** | 태그 추출은 GPT-3.5/Gemini Flash로 충분 |
| **관계도만 고급 모델** | 세션당 1회이므로 GPT-4/Claude 사용 가능 |

---

## 아키텍처 개요

```mermaid
flowchart TB
    subgraph API_Server
        API[API Controller]
        QC[Queue Client]
    end

    subgraph Worker_Server
        WS[Asynq Server]
        TH[Tag Handler]
        MH[Mindmap Handler]
    end

    subgraph AI_Layer
        PM[ProviderManager]
        PM --> OAI[OpenAI Provider]
        PM --> GEM[Gemini Provider]
        PM --> CLA[Claude Provider]
    end

    subgraph Infrastructure
        Redis[(Redis)]
        DB[(PostgreSQL)]
    end

    API --> QC
    QC --> Redis
    Redis --> WS
    WS --> TH
    WS --> MH
    TH --> PM
    MH --> PM
    TH --> DB
    MH --> DB
```

### 설계 원칙

1. **인터페이스 기반 설계**: 모든 AI 프로바이더는 동일한 인터페이스 구현
2. **런타임 프로바이더 전환**: 환경변수 또는 설정으로 프로바이더 변경 가능
3. **Fallback 지원**: 기본 프로바이더 실패 시 대체 프로바이더 사용
4. **용도별 프로바이더 분리**: 태그 추출과 관계도 생성에 다른 모델 사용 가능
5. **비동기 처리**: Asynq Worker를 통한 백그라운드 AI 처리

---

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 10.1 | AI Provider 인터페이스 정의 | ⬜ |
| 10.2 | 개별 Provider 구현 (OpenAI, Gemini, Claude) | ⬜ |
| 10.3 | Provider Manager 및 Config | ⬜ |
| 10.4 | 태그 추출 Worker Handler | ⬜ |
| 10.5 | 마인드맵 생성 Worker Handler | ⬜ |
| 10.6 | UsageService 연동 (토큰 측정) | ⬜ |

---

## Step 10.1: AI Provider 인터페이스 정의

### 체크리스트

- [ ] **공통 타입 정의**
  - [ ] `pkg/infra/ai/types.go`

    ```go
    package ai

    import "context"

    // Message represents a chat message
    type Message struct {
        Role    Role   `json:"role"`
        Content string `json:"content"`
    }

    // Role defines message roles
    type Role string

    const (
        RoleSystem    Role = "system"
        RoleUser      Role = "user"
        RoleAssistant Role = "assistant"
    )

    // ChatOptions contains optional parameters for chat completion
    type ChatOptions struct {
        Temperature   float64  `json:"temperature,omitempty"`
        MaxTokens     int      `json:"max_tokens,omitempty"`
        TopP          float64  `json:"top_p,omitempty"`
        StopSequences []string `json:"stop_sequences,omitempty"`
        JSONMode      bool     `json:"json_mode,omitempty"` // Force JSON output
    }

    // DefaultChatOptions returns sensible defaults
    func DefaultChatOptions() ChatOptions {
        return ChatOptions{
            Temperature: 0.7,
            MaxTokens:   4096,
            TopP:        1.0,
        }
    }

    // ChatResponse contains the AI response
    type ChatResponse struct {
        Content      string `json:"content"`
        Model        string `json:"model"`
        Provider     string `json:"provider"`
        InputTokens  int    `json:"input_tokens,omitempty"`
        OutputTokens int    `json:"output_tokens,omitempty"`
    }

    // ProviderType identifies the AI provider
    type ProviderType string

    const (
        ProviderOpenAI    ProviderType = "openai"
        ProviderGemini    ProviderType = "gemini"
        ProviderClaude    ProviderType = "claude"
    )

    // ProviderConfig holds configuration for a single provider
    type ProviderConfig struct {
        Type     ProviderType `json:"type"`
        APIKey   string       `json:"api_key"`
        Model    string       `json:"model"`
        Enabled  bool         `json:"enabled"`
        Priority int          `json:"priority"` // Lower = higher priority for fallback
    }
    ```

- [ ] **AIProvider 인터페이스 정의**
  - [ ] `pkg/infra/ai/provider.go`

    ```go
    package ai

    import (
        "context"
        "errors"
    )

    var (
        ErrProviderNotConfigured = errors.New("ai provider not configured")
        ErrNoResponse            = errors.New("no response from ai provider")
        ErrRateLimited           = errors.New("rate limited by ai provider")
        ErrInvalidAPIKey         = errors.New("invalid api key")
        ErrContextCanceled       = errors.New("context canceled")
    )

    // AIProvider defines the interface that all AI providers must implement
    type AIProvider interface {
        // Chat sends messages and returns a response
        Chat(ctx context.Context, messages []Message, opts ChatOptions) (*ChatResponse, error)

        // ChatWithJSON is a convenience method that forces JSON output
        ChatWithJSON(ctx context.Context, messages []Message, opts ChatOptions) (*ChatResponse, error)

        // Name returns the provider name for logging/metrics
        Name() string

        // Type returns the provider type
        Type() ProviderType

        // Model returns the current model being used
        Model() string

        // IsHealthy checks if the provider is available
        IsHealthy(ctx context.Context) bool
    }

    // BaseProvider contains common functionality for all providers
    type BaseProvider struct {
        providerType ProviderType
        model        string
    }

    func (b *BaseProvider) Type() ProviderType {
        return b.providerType
    }

    func (b *BaseProvider) Model() string {
        return b.model
    }

    func (b *BaseProvider) Name() string {
        return string(b.providerType)
    }
    ```

### 검증

```bash
cd apps/backend
go build ./...
# 컴파일 성공
```

---

## Step 10.2: 개별 Provider 구현

### 체크리스트

- [ ] **의존성 추가**

  ```bash
  cd apps/backend
  # OpenAI
  go get github.com/sashabaranov/go-openai

  # Google Gemini
  go get github.com/google/generative-ai-go

  # Anthropic Claude
  go get github.com/anthropics/anthropic-sdk-go
  ```

- [ ] **OpenAI Provider 구현**
  - [ ] `pkg/infra/ai/openai.go`

    ```go
    package ai

    import (
        "context"
        "fmt"

        "github.com/sashabaranov/go-openai"
    )

    type OpenAIProvider struct {
        BaseProvider
        client *openai.Client
    }

    func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
        if model == "" {
            model = "gpt-4-turbo-preview"
        }
        return &OpenAIProvider{
            BaseProvider: BaseProvider{
                providerType: ProviderOpenAI,
                model:        model,
            },
            client: openai.NewClient(apiKey),
        }
    }

    func (p *OpenAIProvider) Chat(ctx context.Context, messages []Message, opts ChatOptions) (*ChatResponse, error) {
        chatMessages := make([]openai.ChatCompletionMessage, len(messages))
        for i, msg := range messages {
            chatMessages[i] = openai.ChatCompletionMessage{
                Role:    string(msg.Role),
                Content: msg.Content,
            }
        }

        req := openai.ChatCompletionRequest{
            Model:       p.model,
            Messages:    chatMessages,
            Temperature: float32(opts.Temperature),
            MaxTokens:   opts.MaxTokens,
            TopP:        float32(opts.TopP),
            Stop:        opts.StopSequences,
        }

        resp, err := p.client.CreateChatCompletion(ctx, req)
        if err != nil {
            return nil, fmt.Errorf("openai chat: %w", err)
        }

        if len(resp.Choices) == 0 {
            return nil, ErrNoResponse
        }

        return &ChatResponse{
            Content:      resp.Choices[0].Message.Content,
            Model:        resp.Model,
            Provider:     string(ProviderOpenAI),
            InputTokens:  resp.Usage.PromptTokens,
            OutputTokens: resp.Usage.CompletionTokens,
        }, nil
    }

    func (p *OpenAIProvider) ChatWithJSON(ctx context.Context, messages []Message, opts ChatOptions) (*ChatResponse, error) {
        chatMessages := make([]openai.ChatCompletionMessage, len(messages))
        for i, msg := range messages {
            chatMessages[i] = openai.ChatCompletionMessage{
                Role:    string(msg.Role),
                Content: msg.Content,
            }
        }

        req := openai.ChatCompletionRequest{
            Model:       p.model,
            Messages:    chatMessages,
            Temperature: float32(opts.Temperature),
            MaxTokens:   opts.MaxTokens,
            TopP:        float32(opts.TopP),
            ResponseFormat: &openai.ChatCompletionResponseFormat{
                Type: openai.ChatCompletionResponseFormatTypeJSONObject,
            },
        }

        resp, err := p.client.CreateChatCompletion(ctx, req)
        if err != nil {
            return nil, fmt.Errorf("openai chat json: %w", err)
        }

        if len(resp.Choices) == 0 {
            return nil, ErrNoResponse
        }

        return &ChatResponse{
            Content:      resp.Choices[0].Message.Content,
            Model:        resp.Model,
            Provider:     string(ProviderOpenAI),
            InputTokens:  resp.Usage.PromptTokens,
            OutputTokens: resp.Usage.CompletionTokens,
        }, nil
    }

    func (p *OpenAIProvider) IsHealthy(ctx context.Context) bool {
        _, err := p.Chat(ctx, []Message{
            {Role: RoleUser, Content: "ping"},
        }, ChatOptions{MaxTokens: 5})
        return err == nil
    }
    ```

- [ ] **Google Gemini Provider 구현**
  - [ ] `pkg/infra/ai/gemini.go`

    ```go
    package ai

    import (
        "context"
        "encoding/json"
        "fmt"
        "strings"

        "github.com/google/generative-ai-go/genai"
        "google.golang.org/api/option"
    )

    type GeminiProvider struct {
        BaseProvider
        client *genai.Client
    }

    func NewGeminiProvider(ctx context.Context, apiKey, model string) (*GeminiProvider, error) {
        if model == "" {
            model = "gemini-1.5-pro"
        }

        client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
        if err != nil {
            return nil, fmt.Errorf("create gemini client: %w", err)
        }

        return &GeminiProvider{
            BaseProvider: BaseProvider{
                providerType: ProviderGemini,
                model:        model,
            },
            client: client,
        }, nil
    }

    func (p *GeminiProvider) Chat(ctx context.Context, messages []Message, opts ChatOptions) (*ChatResponse, error) {
        model := p.client.GenerativeModel(p.model)

        model.SetTemperature(float32(opts.Temperature))
        model.SetMaxOutputTokens(int32(opts.MaxTokens))
        model.SetTopP(float32(opts.TopP))

        if len(opts.StopSequences) > 0 {
            model.StopSequences = opts.StopSequences
        }

        var parts []genai.Part
        var systemPrompt string

        for _, msg := range messages {
            switch msg.Role {
            case RoleSystem:
                systemPrompt = msg.Content
            case RoleUser, RoleAssistant:
                parts = append(parts, genai.Text(msg.Content))
            }
        }

        if systemPrompt != "" {
            model.SystemInstruction = &genai.Content{
                Parts: []genai.Part{genai.Text(systemPrompt)},
            }
        }

        resp, err := model.GenerateContent(ctx, parts...)
        if err != nil {
            return nil, fmt.Errorf("gemini generate: %w", err)
        }

        if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
            return nil, ErrNoResponse
        }

        var content strings.Builder
        for _, part := range resp.Candidates[0].Content.Parts {
            if text, ok := part.(genai.Text); ok {
                content.WriteString(string(text))
            }
        }

        return &ChatResponse{
            Content:      content.String(),
            Model:        p.model,
            Provider:     string(ProviderGemini),
            InputTokens:  int(resp.UsageMetadata.PromptTokenCount),
            OutputTokens: int(resp.UsageMetadata.CandidatesTokenCount),
        }, nil
    }

    func (p *GeminiProvider) ChatWithJSON(ctx context.Context, messages []Message, opts ChatOptions) (*ChatResponse, error) {
        model := p.client.GenerativeModel(p.model)

        model.SetTemperature(float32(opts.Temperature))
        model.SetMaxOutputTokens(int32(opts.MaxTokens))
        model.SetTopP(float32(opts.TopP))
        model.ResponseMIMEType = "application/json"

        var parts []genai.Part
        var systemPrompt string

        for _, msg := range messages {
            switch msg.Role {
            case RoleSystem:
                systemPrompt = msg.Content
            case RoleUser, RoleAssistant:
                parts = append(parts, genai.Text(msg.Content))
            }
        }

        if systemPrompt != "" {
            model.SystemInstruction = &genai.Content{
                Parts: []genai.Part{genai.Text(systemPrompt)},
            }
        }

        resp, err := model.GenerateContent(ctx, parts...)
        if err != nil {
            return nil, fmt.Errorf("gemini generate json: %w", err)
        }

        if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
            return nil, ErrNoResponse
        }

        var content strings.Builder
        for _, part := range resp.Candidates[0].Content.Parts {
            if text, ok := part.(genai.Text); ok {
                content.WriteString(string(text))
            }
        }

        var js json.RawMessage
        if err := json.Unmarshal([]byte(content.String()), &js); err != nil {
            return nil, fmt.Errorf("invalid json response: %w", err)
        }

        return &ChatResponse{
            Content:      content.String(),
            Model:        p.model,
            Provider:     string(ProviderGemini),
            InputTokens:  int(resp.UsageMetadata.PromptTokenCount),
            OutputTokens: int(resp.UsageMetadata.CandidatesTokenCount),
        }, nil
    }

    func (p *GeminiProvider) IsHealthy(ctx context.Context) bool {
        _, err := p.Chat(ctx, []Message{
            {Role: RoleUser, Content: "ping"},
        }, ChatOptions{MaxTokens: 5})
        return err == nil
    }

    func (p *GeminiProvider) Close() error {
        return p.client.Close()
    }
    ```

- [ ] **Anthropic Claude Provider 구현**
  - [ ] `pkg/infra/ai/claude.go`

    ```go
    package ai

    import (
        "context"
        "encoding/json"
        "fmt"

        "github.com/anthropics/anthropic-sdk-go"
        "github.com/anthropics/anthropic-sdk-go/option"
    )

    type ClaudeProvider struct {
        BaseProvider
        client *anthropic.Client
    }

    func NewClaudeProvider(apiKey, model string) *ClaudeProvider {
        if model == "" {
            model = "claude-3-5-sonnet-20241022"
        }

        client := anthropic.NewClient(
            option.WithAPIKey(apiKey),
        )

        return &ClaudeProvider{
            BaseProvider: BaseProvider{
                providerType: ProviderClaude,
                model:        model,
            },
            client: client,
        }
    }

    func (p *ClaudeProvider) Chat(ctx context.Context, messages []Message, opts ChatOptions) (*ChatResponse, error) {
        var systemPrompt string
        var anthropicMessages []anthropic.MessageParam

        for _, msg := range messages {
            switch msg.Role {
            case RoleSystem:
                systemPrompt = msg.Content
            case RoleUser:
                anthropicMessages = append(anthropicMessages, anthropic.NewUserMessage(
                    anthropic.NewTextBlock(msg.Content),
                ))
            case RoleAssistant:
                anthropicMessages = append(anthropicMessages, anthropic.NewAssistantMessage(
                    anthropic.NewTextBlock(msg.Content),
                ))
            }
        }

        params := anthropic.MessageNewParams{
            Model:       anthropic.F(p.model),
            MaxTokens:   anthropic.F(int64(opts.MaxTokens)),
            Messages:    anthropic.F(anthropicMessages),
        }

        if systemPrompt != "" {
            params.System = anthropic.F([]anthropic.TextBlockParam{
                anthropic.NewTextBlock(systemPrompt),
            })
        }

        if opts.Temperature > 0 {
            params.Temperature = anthropic.F(opts.Temperature)
        }

        if opts.TopP > 0 && opts.TopP < 1 {
            params.TopP = anthropic.F(opts.TopP)
        }

        if len(opts.StopSequences) > 0 {
            params.StopSequences = anthropic.F(opts.StopSequences)
        }

        resp, err := p.client.Messages.New(ctx, params)
        if err != nil {
            return nil, fmt.Errorf("claude message: %w", err)
        }

        if len(resp.Content) == 0 {
            return nil, ErrNoResponse
        }

        var content string
        for _, block := range resp.Content {
            if block.Type == anthropic.ContentBlockTypeText {
                content += block.Text
            }
        }

        return &ChatResponse{
            Content:      content,
            Model:        string(resp.Model),
            Provider:     string(ProviderClaude),
            InputTokens:  int(resp.Usage.InputTokens),
            OutputTokens: int(resp.Usage.OutputTokens),
        }, nil
    }

    func (p *ClaudeProvider) ChatWithJSON(ctx context.Context, messages []Message, opts ChatOptions) (*ChatResponse, error) {
        modifiedMessages := make([]Message, len(messages))
        copy(modifiedMessages, messages)

        for i := len(modifiedMessages) - 1; i >= 0; i-- {
            if modifiedMessages[i].Role == RoleUser {
                modifiedMessages[i].Content += "\n\nRespond with valid JSON only. No markdown, no explanation."
                break
            }
        }

        resp, err := p.Chat(ctx, modifiedMessages, opts)
        if err != nil {
            return nil, err
        }

        var js json.RawMessage
        if err := json.Unmarshal([]byte(resp.Content), &js); err != nil {
            return nil, fmt.Errorf("invalid json response: %w", err)
        }

        return resp, nil
    }

    func (p *ClaudeProvider) IsHealthy(ctx context.Context) bool {
        _, err := p.Chat(ctx, []Message{
            {Role: RoleUser, Content: "ping"},
        }, ChatOptions{MaxTokens: 5})
        return err == nil
    }
    ```

### 검증

```bash
cd apps/backend
go build ./...
# 컴파일 성공
```

---

## Step 10.3: Provider Manager 및 Config

### 체크리스트

- [ ] **환경 변수 설정**

  ```env
  # AI Provider 설정
  AI_DEFAULT_PROVIDER=openai
  AI_FALLBACK_PROVIDERS=gemini,claude

  # 용도별 프로바이더 (선택적)
  AI_TAG_EXTRACTION_PROVIDER=gemini
  AI_MINDMAP_PROVIDER=openai

  # OpenAI
  OPENAI_API_KEY=sk-...
  OPENAI_MODEL=gpt-4-turbo-preview

  # Google Gemini
  GEMINI_API_KEY=...
  GEMINI_MODEL=gemini-1.5-flash

  # Anthropic Claude
  CLAUDE_API_KEY=sk-ant-...
  CLAUDE_MODEL=claude-3-5-sonnet-20241022
  ```

- [ ] **Config 업데이트**
  - [ ] `pkg/config/config.go`에 AI 설정 추가

    ```go
    type Config struct {
        // ... 기존 필드
        AI AIConfig
    }

    type AIConfig struct {
        DefaultProvider      string   `json:"default_provider"`
        FallbackProviders    []string `json:"fallback_providers"`
        TagExtractionProvider string  `json:"tag_extraction_provider"`
        MindmapProvider      string   `json:"mindmap_provider"`

        OpenAI  OpenAIConfig  `json:"openai"`
        Gemini  GeminiConfig  `json:"gemini"`
        Claude  ClaudeConfig  `json:"claude"`
    }

    type OpenAIConfig struct {
        APIKey  string `json:"api_key"`
        Model   string `json:"model"`
        Enabled bool   `json:"enabled"`
    }

    type GeminiConfig struct {
        APIKey  string `json:"api_key"`
        Model   string `json:"model"`
        Enabled bool   `json:"enabled"`
    }

    type ClaudeConfig struct {
        APIKey  string `json:"api_key"`
        Model   string `json:"model"`
        Enabled bool   `json:"enabled"`
    }
    ```

- [ ] **Provider Manager 구현**
  - [ ] `pkg/infra/ai/manager.go`

    ```go
    package ai

    import (
        "context"
        "fmt"
        "log/slog"
        "sync"
    )

    // TaskType identifies the AI task for provider selection
    type TaskType string

    const (
        TaskTagExtraction TaskType = "tag_extraction"
        TaskMindmap       TaskType = "mindmap"
        TaskGeneral       TaskType = "general"
    )

    // ProviderManager manages multiple AI providers with fallback support
    type ProviderManager struct {
        providers       map[ProviderType]AIProvider
        defaultProvider ProviderType
        fallbackOrder   []ProviderType
        taskProviders   map[TaskType]ProviderType
        mu              sync.RWMutex
    }

    // NewProviderManager creates a new provider manager from config
    func NewProviderManager(ctx context.Context, cfg AIConfig) (*ProviderManager, error) {
        pm := &ProviderManager{
            providers:     make(map[ProviderType]AIProvider),
            taskProviders: make(map[TaskType]ProviderType),
        }

        // Initialize enabled providers
        if cfg.OpenAI.Enabled {
            pm.providers[ProviderOpenAI] = NewOpenAIProvider(cfg.OpenAI.APIKey, cfg.OpenAI.Model)
            slog.Info("initialized ai provider", "provider", "openai", "model", cfg.OpenAI.Model)
        }

        if cfg.Gemini.Enabled {
            gemini, err := NewGeminiProvider(ctx, cfg.Gemini.APIKey, cfg.Gemini.Model)
            if err != nil {
                slog.Warn("failed to initialize gemini provider", "error", err)
            } else {
                pm.providers[ProviderGemini] = gemini
                slog.Info("initialized ai provider", "provider", "gemini", "model", cfg.Gemini.Model)
            }
        }

        if cfg.Claude.Enabled {
            pm.providers[ProviderClaude] = NewClaudeProvider(cfg.Claude.APIKey, cfg.Claude.Model)
            slog.Info("initialized ai provider", "provider", "claude", "model", cfg.Claude.Model)
        }

        if len(pm.providers) == 0 {
            return nil, fmt.Errorf("no ai providers configured")
        }

        // Set default provider
        pm.defaultProvider = ProviderType(cfg.DefaultProvider)
        if _, ok := pm.providers[pm.defaultProvider]; !ok {
            for pt := range pm.providers {
                pm.defaultProvider = pt
                break
            }
        }

        // Set fallback order
        for _, name := range cfg.FallbackProviders {
            pt := ProviderType(name)
            if _, ok := pm.providers[pt]; ok {
                pm.fallbackOrder = append(pm.fallbackOrder, pt)
            }
        }

        // Set task-specific providers
        if cfg.TagExtractionProvider != "" {
            pt := ProviderType(cfg.TagExtractionProvider)
            if _, ok := pm.providers[pt]; ok {
                pm.taskProviders[TaskTagExtraction] = pt
            }
        }

        if cfg.MindmapProvider != "" {
            pt := ProviderType(cfg.MindmapProvider)
            if _, ok := pm.providers[pt]; ok {
                pm.taskProviders[TaskMindmap] = pt
            }
        }

        slog.Info("provider manager initialized",
            "default", pm.defaultProvider,
            "fallbacks", pm.fallbackOrder,
            "task_providers", pm.taskProviders,
        )

        return pm, nil
    }

    // ChatWithJSON executes chat with JSON mode and automatic fallback
    func (pm *ProviderManager) ChatWithJSON(ctx context.Context, task TaskType, messages []Message, opts ChatOptions) (*ChatResponse, error) {
        return pm.chatWithFallback(ctx, task, messages, opts, true)
    }

    func (pm *ProviderManager) chatWithFallback(ctx context.Context, task TaskType, messages []Message, opts ChatOptions, jsonMode bool) (*ChatResponse, error) {
        pm.mu.RLock()
        providers := pm.getProvidersInOrder(task)
        pm.mu.RUnlock()

        var lastErr error
        for _, provider := range providers {
            slog.Debug("attempting ai request",
                "provider", provider.Name(),
                "model", provider.Model(),
                "task", task,
            )

            var resp *ChatResponse
            var err error

            if jsonMode {
                resp, err = provider.ChatWithJSON(ctx, messages, opts)
            } else {
                resp, err = provider.Chat(ctx, messages, opts)
            }

            if err == nil {
                slog.Info("ai request successful",
                    "provider", provider.Name(),
                    "model", resp.Model,
                    "input_tokens", resp.InputTokens,
                    "output_tokens", resp.OutputTokens,
                )
                return resp, nil
            }

            lastErr = err
            slog.Warn("ai provider failed, trying fallback",
                "provider", provider.Name(),
                "error", err,
            )
        }

        return nil, fmt.Errorf("all ai providers failed, last error: %w", lastErr)
    }

    func (pm *ProviderManager) getProvidersInOrder(task TaskType) []AIProvider {
        var result []AIProvider
        seen := make(map[ProviderType]bool)

        if pt, ok := pm.taskProviders[task]; ok {
            if p, ok := pm.providers[pt]; ok {
                result = append(result, p)
                seen[pt] = true
            }
        }

        if !seen[pm.defaultProvider] {
            if p, ok := pm.providers[pm.defaultProvider]; ok {
                result = append(result, p)
                seen[pm.defaultProvider] = true
            }
        }

        for _, pt := range pm.fallbackOrder {
            if !seen[pt] {
                if p, ok := pm.providers[pt]; ok {
                    result = append(result, p)
                    seen[pt] = true
                }
            }
        }

        return result
    }

    // Close closes all providers
    func (pm *ProviderManager) Close() error {
        pm.mu.Lock()
        defer pm.mu.Unlock()

        for _, provider := range pm.providers {
            if closer, ok := provider.(interface{ Close() error }); ok {
                if err := closer.Close(); err != nil {
                    slog.Warn("failed to close provider", "provider", provider.Name(), "error", err)
                }
            }
        }
        return nil
    }
    ```

### 검증

```bash
cd apps/backend
go build ./...
# 컴파일 성공
```

---

## Step 10.4: 태그 추출 Worker Handler

### 목표

이벤트 배치 수신 시 새로운 URL에 대해 Asynq Task를 생성하고, Worker에서 태그를 추출합니다.

### 체크리스트

- [ ] **태그 추출 Task 정의 (Phase 6에서 이미 정의됨)**
  - [ ] `pkg/infra/queue/tasks.go`에 추가 확인

    ```go
    const TypeURLTagExtraction = "url:tag_extraction"

    type URLTagExtractionPayload struct {
        URLID string `json:"url_id"`
    }

    func NewURLTagExtractionTask(urlID string) (*asynq.Task, error) {
        payload, err := json.Marshal(URLTagExtractionPayload{URLID: urlID})
        if err != nil {
            return nil, err
        }
        return asynq.NewTask(TypeURLTagExtraction, payload), nil
    }
    ```

- [ ] **태그 추출 Handler 구현**
  - [ ] `internal/worker/handler/tag_extraction.go`

    ```go
    package handler

    import (
        "context"
        "encoding/json"
        "fmt"
        "log/slog"

        "github.com/google/uuid"
        "github.com/hibiken/asynq"

        "github.com/mindhit/api/pkg/ent"
        "github.com/mindhit/api/pkg/infra/ai"
        "github.com/mindhit/api/pkg/infra/queue"
    )

    const tagExtractionPrompt = `웹 페이지를 분석하고 다음을 추출하세요:

1. 핵심 태그 3-5개 (한국어, 명사형)
2. 1-2문장 요약 (한국어)

페이지 제목: %s
페이지 내용:
%s

JSON 형식으로 응답:
{
  "tags": ["태그1", "태그2", "태그3"],
  "summary": "페이지 요약"
}`

    type TagResult struct {
        Tags    []string `json:"tags"`
        Summary string   `json:"summary"`
    }

    func (h *handlers) HandleURLTagExtraction(ctx context.Context, t *asynq.Task) error {
        var payload queue.URLTagExtractionPayload
        if err := json.Unmarshal(t.Payload(), &payload); err != nil {
            return fmt.Errorf("unmarshal payload: %w", err)
        }

        urlID, err := uuid.Parse(payload.URLID)
        if err != nil {
            return fmt.Errorf("parse url id: %w", err)
        }

        slog.Info("extracting tags", "url_id", payload.URLID)

        // Get URL from database
        u, err := h.client.URL.Get(ctx, urlID)
        if err != nil {
            return fmt.Errorf("get url: %w", err)
        }

        // Skip if already has tags
        if len(u.Tags) > 0 {
            slog.Debug("url already has tags, skipping", "url", u.URL)
            return nil
        }

        // Skip if no content
        if u.Content == "" {
            slog.Warn("url has no content, skipping", "url", u.URL)
            return nil
        }

        // Generate tags using AI
        messages := []ai.Message{
            {
                Role:    ai.RoleUser,
                Content: fmt.Sprintf(tagExtractionPrompt, u.Title, truncateContent(u.Content, 8000)),
            },
        }

        opts := ai.DefaultChatOptions()
        opts.MaxTokens = 500

        response, err := h.aiManager.ChatWithJSON(ctx, ai.TaskTagExtraction, messages, opts)
        if err != nil {
            return fmt.Errorf("ai tag extraction: %w", err)
        }

        var result TagResult
        if err := json.Unmarshal([]byte(response.Content), &result); err != nil {
            return fmt.Errorf("parse ai response: %w", err)
        }

        // Update URL with tags and summary
        _, err = h.client.URL.UpdateOneID(urlID).
            SetTags(result.Tags).
            SetSummary(result.Summary).
            Save(ctx)

        if err != nil {
            return fmt.Errorf("update url: %w", err)
        }

        slog.Info("extracted tags",
            "url", u.URL,
            "tags", result.Tags,
            "provider", response.Provider,
            "tokens", response.InputTokens+response.OutputTokens,
        )
        return nil
    }

    func truncateContent(content string, maxLen int) string {
        if len(content) <= maxLen {
            return content
        }
        return content[:maxLen] + "..."
    }
    ```

- [ ] **Handler 등록 업데이트**
  - [ ] `internal/worker/handler/handler.go`

    ```go
    package handler

    import (
        "github.com/mindhit/api/pkg/ent"
        "github.com/mindhit/api/pkg/infra/ai"
        "github.com/mindhit/api/pkg/infra/queue"
    )

    type handlers struct {
        client    *ent.Client
        aiManager *ai.ProviderManager
    }

    func RegisterHandlers(server *queue.Server, client *ent.Client, aiManager *ai.ProviderManager) {
        h := &handlers{
            client:    client,
            aiManager: aiManager,
        }

        server.HandleFunc(queue.TypeSessionProcess, h.HandleSessionProcess)
        server.HandleFunc(queue.TypeSessionCleanup, h.HandleSessionCleanup)
        server.HandleFunc(queue.TypeURLTagExtraction, h.HandleURLTagExtraction)
        server.HandleFunc(queue.TypeMindmapGenerate, h.HandleMindmapGenerate)
    }
    ```

- [ ] **API에서 새 URL 발견 시 Task Enqueue**
  - [ ] `pkg/service/event_service.go`

    ```go
    type EventService struct {
        client      *ent.Client
        queueClient *queue.Client
    }

    func (s *EventService) ProcessBatchEvents(ctx context.Context, sessionID uuid.UUID, events []Event) error {
        for _, event := range events {
            if event.Type == "page_visit" {
                urlID, isNew, err := s.upsertURL(ctx, event)
                if err != nil {
                    slog.Error("failed to upsert url", "error", err)
                    continue
                }

                // Enqueue tag extraction for new URLs
                if isNew {
                    task, err := queue.NewURLTagExtractionTask(urlID.String())
                    if err != nil {
                        slog.Error("failed to create tag extraction task", "error", err)
                        continue
                    }

                    _, err = s.queueClient.Enqueue(task, asynq.MaxRetry(3))
                    if err != nil {
                        slog.Error("failed to enqueue tag extraction task", "error", err)
                    }
                }

                // Save page visit...
            }
        }
        return nil
    }
    ```

### 검증

```bash
# Worker 실행
cd apps/backend
REDIS_ADDR=localhost:6379 go run ./cmd/worker

# API 실행 후 이벤트 전송
curl -X POST http://localhost:8080/api/v1/events/batch \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"session_id": "...", "events": [{"type": "page_visit", ...}]}'

# Worker 로그 확인
# "extracting tags" url_id=...
# "extracted tags" url=https://... tags=["AI", "머신러닝"]
```

---

## Step 10.5: 마인드맵 생성 Worker Handler

### 목표

세션 종료 시 Asynq Task를 생성하고, Worker에서 관계도 JSON을 생성합니다.

### 체크리스트

- [ ] **마인드맵 타입 정의**
  - [ ] `pkg/service/mindmap_types.go`

    ```go
    package service

    type MindmapNode struct {
        ID       string                 `json:"id"`
        Label    string                 `json:"label"`
        Type     string                 `json:"type"` // core, topic, subtopic, page
        Size     float64                `json:"size"`
        Color    string                 `json:"color"`
        Position *Position              `json:"position,omitempty"`
        Data     map[string]interface{} `json:"data"`
    }

    type Position struct {
        X float64 `json:"x"`
        Y float64 `json:"y"`
        Z float64 `json:"z"`
    }

    type MindmapEdge struct {
        Source string  `json:"source"`
        Target string  `json:"target"`
        Weight float64 `json:"weight"`
        Label  string  `json:"label,omitempty"`
    }

    type MindmapLayout struct {
        Type   string                 `json:"type"` // galaxy, tree, radial
        Params map[string]interface{} `json:"params"`
    }

    type MindmapData struct {
        Nodes  []MindmapNode `json:"nodes"`
        Edges  []MindmapEdge `json:"edges"`
        Layout MindmapLayout `json:"layout"`
    }
    ```

- [ ] **마인드맵 생성 Handler 구현**
  - [ ] `internal/worker/handler/mindmap.go`

    ```go
    package handler

    import (
        "context"
        "encoding/json"
        "fmt"
        "log/slog"
        "math"
        "strings"

        "github.com/google/uuid"
        "github.com/hibiken/asynq"

        "github.com/mindhit/api/pkg/ent"
        "github.com/mindhit/api/pkg/ent/session"
        "github.com/mindhit/api/pkg/infra/ai"
        "github.com/mindhit/api/pkg/infra/queue"
        "github.com/mindhit/api/pkg/service"
    )

    const relationshipGraphPrompt = `브라우징 세션의 페이지들과 추출된 태그를 분석하여 관계도를 생성하세요.

## 세션 데이터

### 방문한 페이지들 (URL + 태그 + 요약)

%s

### 하이라이트 (사용자가 선택한 텍스트)

%s

## 요청사항

1. **핵심 주제 (core)**: 세션 전체를 관통하는 중심 테마 1개
2. **주요 토픽 (topics)**: 공통 태그를 기반으로 3-5개 그룹화
3. **페이지 연결**: 각 토픽에 해당하는 페이지들 매핑
4. **토픽 간 연결 (connections)**: 태그가 겹치는 토픽들의 관계

## JSON 형식으로 응답

{
  "core": {
    "label": "핵심 주제 (한국어)",
    "description": "세션 전체 요약 (1-2문장)"
  },
  "topics": [
    {
      "id": "topic-1",
      "label": "토픽명 (한국어)",
      "tags": ["관련", "태그들"],
      "description": "토픽 설명",
      "pages": [
        {
          "url_id": "uuid",
          "title": "페이지 제목",
          "relevance": 0.9
        }
      ]
    }
  ],
  "connections": [
    {
      "from": "topic-1",
      "to": "topic-2",
      "shared_tags": ["공통태그"],
      "reason": "연결 이유"
    }
  ]
}`

    type RelationshipGraphResponse struct {
        Core struct {
            Label       string `json:"label"`
            Description string `json:"description"`
        } `json:"core"`
        Topics []struct {
            ID          string   `json:"id"`
            Label       string   `json:"label"`
            Tags        []string `json:"tags"`
            Description string   `json:"description"`
            Pages       []struct {
                URLID     string  `json:"url_id"`
                Title     string  `json:"title"`
                Relevance float64 `json:"relevance"`
            } `json:"pages"`
        } `json:"topics"`
        Connections []struct {
            From       string   `json:"from"`
            To         string   `json:"to"`
            SharedTags []string `json:"shared_tags"`
            Reason     string   `json:"reason"`
        } `json:"connections"`
    }

    func (h *handlers) HandleMindmapGenerate(ctx context.Context, t *asynq.Task) error {
        var payload queue.MindmapGeneratePayload
        if err := json.Unmarshal(t.Payload(), &payload); err != nil {
            return fmt.Errorf("unmarshal payload: %w", err)
        }

        sessionID, err := uuid.Parse(payload.SessionID)
        if err != nil {
            return fmt.Errorf("parse session id: %w", err)
        }

        slog.Info("generating mindmap", "session_id", payload.SessionID)

        // Get session with all related data
        sess, err := h.client.Session.
            Query().
            Where(session.IDEQ(sessionID)).
            WithPageVisits(func(q *ent.PageVisitQuery) {
                q.WithURL()
            }).
            WithHighlights().
            Only(ctx)

        if err != nil {
            return fmt.Errorf("get session: %w", err)
        }

        // Build page data with tags
        var pageData strings.Builder
        dwellTimeMap := make(map[string]int)

        for _, pv := range sess.Edges.PageVisits {
            if pv.Edges.URL == nil {
                continue
            }
            u := pv.Edges.URL

            dwellTime := 0
            if pv.DwellTimeSeconds != nil {
                dwellTime = *pv.DwellTimeSeconds
            }
            dwellTimeMap[u.ID.String()] = dwellTime

            pageData.WriteString(fmt.Sprintf(`

- ID: %s
  제목: %s
  URL: %s
  태그: [%s]
  요약: %s
  체류시간: %d초
`,
                u.ID.String(),
                u.Title,
                u.URL,
                strings.Join(u.Tags, ", "),
                u.Summary,
                dwellTime,
            ))
        }

        // Build highlights text
        var highlights strings.Builder
        if len(sess.Edges.Highlights) > 0 {
            for _, h := range sess.Edges.Highlights {
                highlights.WriteString(fmt.Sprintf("- \"%s\"\n", h.Text))
            }
        } else {
            highlights.WriteString("(하이라이트 없음)")
        }

        // Generate relationship graph using AI
        messages := []ai.Message{
            {
                Role:    ai.RoleUser,
                Content: fmt.Sprintf(relationshipGraphPrompt, pageData.String(), highlights.String()),
            },
        }

        opts := ai.DefaultChatOptions()
        opts.MaxTokens = 4096

        response, err := h.aiManager.ChatWithJSON(ctx, ai.TaskMindmap, messages, opts)
        if err != nil {
            return fmt.Errorf("ai generate mindmap: %w", err)
        }

        var aiResp RelationshipGraphResponse
        if err := json.Unmarshal([]byte(response.Content), &aiResp); err != nil {
            return fmt.Errorf("parse ai response: %w", err)
        }

        // Convert AI response to mindmap data
        mindmapData := buildMindmapFromRelationship(aiResp, dwellTimeMap)

        // Convert to JSON for storage
        nodesJSON, _ := json.Marshal(mindmapData.Nodes)
        edgesJSON, _ := json.Marshal(mindmapData.Edges)
        layoutJSON, _ := json.Marshal(mindmapData.Layout)

        // Save mindmap to database
        _, err = h.client.MindmapGraph.
            Create().
            SetSessionID(sessionID).
            SetNodes(nodesJSON).
            SetEdges(edgesJSON).
            SetLayout(layoutJSON).
            Save(ctx)

        if err != nil {
            return fmt.Errorf("save mindmap: %w", err)
        }

        // Update session status to completed
        _, err = h.client.Session.
            UpdateOneID(sessionID).
            SetStatus(session.StatusCompleted).
            Save(ctx)

        if err != nil {
            return fmt.Errorf("update session status: %w", err)
        }

        slog.Info("mindmap generated",
            "session_id", payload.SessionID,
            "topics", len(aiResp.Topics),
            "connections", len(aiResp.Connections),
            "provider", response.Provider,
        )
        return nil
    }

    func buildMindmapFromRelationship(resp RelationshipGraphResponse, dwellTimeMap map[string]int) service.MindmapData {
        var nodes []service.MindmapNode
        var edges []service.MindmapEdge

        // Core node
        coreID := "core"
        nodes = append(nodes, service.MindmapNode{
            ID:       coreID,
            Label:    resp.Core.Label,
            Type:     "core",
            Size:     100,
            Color:    "#FFD700",
            Position: &service.Position{X: 0, Y: 0, Z: 0},
            Data: map[string]interface{}{
                "description": resp.Core.Description,
            },
        })

        // Topic nodes
        topicCount := len(resp.Topics)
        for i, topic := range resp.Topics {
            topicID := topic.ID
            if topicID == "" {
                topicID = fmt.Sprintf("topic-%d", i)
            }

            angle := (float64(i) / float64(topicCount)) * 2 * math.Pi
            radius := 200.0
            topicSize := 40.0 + float64(len(topic.Pages))*10
            if topicSize > 80 {
                topicSize = 80
            }

            nodes = append(nodes, service.MindmapNode{
                ID:    topicID,
                Label: topic.Label,
                Type:  "topic",
                Size:  topicSize,
                Color: getTopicColor(i),
                Position: &service.Position{
                    X: radius * math.Cos(angle),
                    Y: radius * math.Sin(angle),
                    Z: 0,
                },
                Data: map[string]interface{}{
                    "description": topic.Description,
                    "tags":        topic.Tags,
                },
            })

            edges = append(edges, service.MindmapEdge{
                Source: coreID,
                Target: topicID,
                Weight: 1.0,
            })

            // Page nodes
            for j, page := range topic.Pages {
                pageID := page.URLID
                subAngle := angle + (float64(j)-float64(len(topic.Pages))/2)*0.4
                subRadius := 60.0 + float64(j)*15

                size := 15.0
                if dwell, ok := dwellTimeMap[page.URLID]; ok {
                    size = math.Min(40, 15+float64(dwell)/20)
                }
                size = size * (0.5 + page.Relevance*0.5)

                nodes = append(nodes, service.MindmapNode{
                    ID:    pageID,
                    Label: page.Title,
                    Type:  "page",
                    Size:  size,
                    Color: getTopicColor(i),
                    Position: &service.Position{
                        X: radius*math.Cos(angle) + subRadius*math.Cos(subAngle),
                        Y: radius*math.Sin(angle) + subRadius*math.Sin(subAngle),
                        Z: 0,
                    },
                    Data: map[string]interface{}{
                        "url_id":    page.URLID,
                        "relevance": page.Relevance,
                    },
                })

                edges = append(edges, service.MindmapEdge{
                    Source: topicID,
                    Target: pageID,
                    Weight: page.Relevance,
                })
            }
        }

        // Cross-topic connections
        for _, conn := range resp.Connections {
            edges = append(edges, service.MindmapEdge{
                Source: conn.From,
                Target: conn.To,
                Weight: float64(len(conn.SharedTags)) * 0.2,
                Label:  conn.Reason,
            })
        }

        return service.MindmapData{
            Nodes: nodes,
            Edges: edges,
            Layout: service.MindmapLayout{
                Type: "galaxy",
                Params: map[string]interface{}{
                    "center": []float64{0, 0, 0},
                    "scale":  1.0,
                },
            },
        }
    }

    func getTopicColor(index int) string {
        colors := []string{
            "#3B82F6", "#10B981", "#F59E0B", "#EF4444",
            "#8B5CF6", "#EC4899", "#14B8A6", "#F97316",
        }
        return colors[index%len(colors)]
    }

    ```

- [ ] **세션 Stop 시 Task Enqueue**
  - [ ] `pkg/service/session_service.go`

    ```go
    func (s *SessionService) StopSession(ctx context.Context, sessionID string) (*ent.Session, error) {
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

        // Enqueue mindmap generation task
        task, err := queue.NewMindmapGenerateTask(sessionID)
        if err != nil {
            slog.Error("failed to create mindmap task", "error", err)
            return sess, nil
        }

        _, err = s.queueClient.Enqueue(task, asynq.MaxRetry(3))
        if err != nil {
            slog.Error("failed to enqueue mindmap task", "error", err)
            return sess, nil
        }

        slog.Info("mindmap generation enqueued", "session_id", sessionID)
        return sess, nil
    }
    ```

- [ ] **Worker main.go 업데이트**
  - [ ] `cmd/worker/main.go`

    ```go
    func main() {
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

        // Initialize AI Provider Manager
        ctx := context.Background()
        aiManager, err := ai.NewProviderManager(ctx, cfg.AI)
        if err != nil {
            slog.Error("failed to initialize ai manager", "error", err)
            os.Exit(1)
        }
        defer aiManager.Close()

        // Create worker server
        server := queue.NewServer(queue.ServerConfig{
            RedisAddr:   cfg.RedisAddr,
            Concurrency: cfg.WorkerConcurrency,
        })

        // Register handlers with AI manager
        handler.RegisterHandlers(server, client, aiManager)

        // Graceful shutdown
        go func() {
            sigCh := make(chan os.Signal, 1)
            signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
            <-sigCh
            slog.Info("shutting down worker")
            server.Shutdown()
        }()

        if err := server.Run(); err != nil {
            slog.Error("worker server error", "error", err)
            os.Exit(1)
        }
    }
    ```

### 검증

```bash
# Worker 실행
cd apps/backend
REDIS_ADDR=localhost:6379 \
OPENAI_API_KEY=sk-... \
go run ./cmd/worker

# API에서 세션 종료
curl -X POST http://localhost:8080/api/v1/sessions/{id}/stop \
  -H "Authorization: Bearer $TOKEN"

# Worker 로그 확인
# "generating mindmap" session_id=...
# "mindmap generated" session_id=... topics=4 connections=2 provider=openai
```

---

## 프로바이더 비교

| 프로바이더 | 모델 | 장점 | 단점 | 권장 용도 |
|-----------|------|------|------|----------|
| **OpenAI** | gpt-4-turbo | JSON mode 네이티브 지원, 안정적 | 비용 높음 | 마인드맵 생성 |
| **Gemini** | gemini-1.5-flash | 빠르고 저렴, 긴 컨텍스트 | JSON 불안정 가능 | 태그 추출 |
| **Claude** | claude-3-5-sonnet | 뛰어난 분석력, 한국어 우수 | JSON mode 없음 | 복잡한 분석 |

### 권장 설정

```env
# 비용 최적화 설정
AI_DEFAULT_PROVIDER=gemini
AI_TAG_EXTRACTION_PROVIDER=gemini   # 저비용, 빠름
AI_MINDMAP_PROVIDER=openai          # JSON 안정성
GEMINI_MODEL=gemini-1.5-flash
```

---

## Step 10.6: UsageService 연동 (토큰 측정)

### 목표

AI 서비스에서 토큰 사용량을 **정확하게** 추적하고 제한을 체크합니다.

### 토큰 측정 원리

AI API는 응답에 **실제 사용된 토큰 수**를 포함하여 반환합니다. 예상치가 아닌 정확한 값입니다.

**AI 제공업체별 응답 필드:**

| 제공업체 | 응답 필드 | 예시 값 |
|---------|----------|--------|
| OpenAI | `response.Usage.TotalTokens` | 3847 |
| Google Gemini | `response.UsageMetadata.TotalTokenCount` | 3847 |
| Anthropic Claude | `response.Usage.InputTokens + OutputTokens` | 3847 |

### 체크리스트

- [ ] AI 서비스에 UsageService 의존성 주입
- [ ] 요청 전 사용량 제한 체크
- [ ] API 호출 후 **응답에서 토큰 사용량 추출**
- [ ] token_usage 테이블에 정확한 값 기록
- [ ] 제한 초과 시 적절한 에러 반환

### 코드 예시

```go
// internal/worker/handler/mindmap.go
func (h *handlers) HandleMindmapGenerate(ctx context.Context, t *asynq.Task) error {
    // ... 기존 코드 ...

    // 1. 제한 체크
    status, err := h.usageService.CheckLimit(ctx, userID)
    if err != nil {
        return err
    }
    if !status.CanUseAI {
        return ErrTokenLimitExceeded
    }

    // 2. AI 호출
    response, err := h.aiManager.ChatWithJSON(ctx, ai.TaskMindmap, messages, opts)
    if err != nil {
        return err
    }

    // 3. 토큰 사용량 기록
    tokensUsed := response.InputTokens + response.OutputTokens
    if err := h.usageService.RecordUsage(ctx, service.UsageRecord{
        UserID:    userID,
        SessionID: sessionID,
        Operation: "mindmap",
        Tokens:    tokensUsed,
        AIModel:   response.Model,
    }); err != nil {
        slog.Error("failed to record usage", "error", err)
    }

    // ... 나머지 처리 ...
}
```

---

## Phase 10 완료 확인

### 전체 검증 체크리스트

- [ ] AI Provider 인터페이스 정의
- [ ] OpenAI Provider 구현
- [ ] Google Gemini Provider 구현
- [ ] Anthropic Claude Provider 구현
- [ ] Provider Manager 구현 (Fallback 지원)
- [ ] 태그 추출 Worker Handler
- [ ] 마인드맵 생성 Worker Handler
- [ ] API에서 Task Enqueue 연동
- [ ] Worker에서 AI 처리 완료
- [ ] **UsageService 연동 (토큰 측정)**

### 테스트 요구사항

| 테스트 유형 | 대상 | 파일 |
| ----------- | ---- | ---- |
| 단위 테스트 | Provider Manager Fallback | `ai/manager_test.go` |
| 단위 테스트 | 마인드맵 그래프 생성 | `handler/mindmap_test.go` |
| Mock 테스트 | AI Provider (Mock 응답) | `ai/mock_provider_test.go` |
| 단위 테스트 | 토큰 사용량 기록 | `service/usage_service_test.go` |

```bash
# Phase 10 테스트 실행
moonx backend:test -- -run "TestAI|TestMindmap|TestUsage"
```

> **Note**: AI Provider 테스트는 실제 API 호출 대신 Mock을 사용합니다. 통합 테스트는 별도 환경에서 수동 실행합니다.

### 산출물 요약

| 항목 | 위치 |
| ---- | ---- |
| AI 타입 정의 | `pkg/infra/ai/types.go` |
| Provider 인터페이스 | `pkg/infra/ai/provider.go` |
| OpenAI Provider | `pkg/infra/ai/openai.go` |
| Gemini Provider | `pkg/infra/ai/gemini.go` |
| Claude Provider | `pkg/infra/ai/claude.go` |
| Provider Manager | `pkg/infra/ai/manager.go` |
| 태그 추출 Handler | `internal/worker/handler/tag_extraction.go` |
| 마인드맵 Handler | `internal/worker/handler/mindmap.go` |
| 마인드맵 타입 | `pkg/service/mindmap_types.go` |
| 테스트 | `pkg/infra/ai/*_test.go` |

---

## 다음 Phase

Phase 10 완료 후 [Phase 11: 웹앱 대시보드](./phase-11-dashboard.md)으로 진행하세요.
