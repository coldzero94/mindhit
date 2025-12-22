# Phase 9: AI 마인드맵 생성

## 개요

| 항목 | 내용 |
|-----|------|
| **목표** | 다중 AI 프로바이더(OpenAI, Google Gemini, Anthropic Claude)를 지원하는 태그 추출 및 마인드맵 생성 |
| **선행 조건** | Phase 6 완료 (스케줄러) |
| **예상 소요** | 5 Steps |
| **결과물** | 페이지 방문 시 태그 추출, 세션 종료 시 관계도 JSON 생성 |

---

## 마인드맵 생성 알고리즘

### 핵심 원칙

1. **페이지 방문 시**: LLM으로 태그/키워드 추출 (페이지당 1회, 중복 URL은 재사용)
2. **세션 종료 시**: 추출된 태그들을 기반으로 LLM이 관계도 JSON 생성 (세션당 1회)

### 처리 흐름

```mermaid
sequenceDiagram
    participant EXT as Extension
    participant API as API Server
    participant DB as Database
    participant AI as AI Provider

    Note over EXT,AI: 1. 페이지 방문 시 (실시간)
    EXT->>API: 이벤트 배치 전송 (URL, content)
    API->>DB: URL 중복 체크 (url_hash)
    alt URL이 새로운 경우
        API->>AI: 태그 추출 요청 (content)
        AI-->>API: { tags: [...], summary: "..." }
        API->>DB: urls 테이블에 tags, summary 저장
    else URL이 이미 존재
        API->>DB: 기존 tags 재사용
    end

    Note over EXT,AI: 2. 세션 종료 시 (1회)
    EXT->>API: 세션 Stop 요청
    API->>DB: 세션 상태 → processing
    API->>DB: 세션의 모든 URL + tags 조회
    API->>AI: 관계도 생성 요청 (tags 목록)
    AI-->>API: { nodes: [...], edges: [...] }
    API->>DB: mindmap_graphs 저장
    API->>DB: 세션 상태 → completed
```

### 태그 추출 (페이지당)

| 항목 | 설명 |
|-----|------|
| **트리거** | 이벤트 배치 수신 시 새로운 URL 감지 |
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
| **트리거** | 세션 종료 (Stop) 시 |
| **입력** | 세션의 모든 URL + tags + 체류시간 + 하이라이트 |
| **출력** | 마인드맵 JSON (nodes, edges) |
| **저장** | `mindmap_graphs` 테이블 |

**프롬프트 예시:**
```
브라우징 세션 데이터를 분석하고 마인드맵 구조를 생성하세요.

세션 데이터:
- URL 1: [태그: AI, 머신러닝] 체류시간: 5분
- URL 2: [태그: 딥러닝, 신경망] 체류시간: 3분
- URL 3: [태그: AI, 자연어처리] 체류시간: 8분
- 하이라이트: "GPT-4는 가장 강력한..."

다음 JSON 형식으로 응답:
{
  "core": { "label": "핵심 주제" },
  "topics": [
    {
      "label": "주요 토픽",
      "subtopics": [
        { "label": "하위 토픽", "url_ids": ["uuid1"] }
      ]
    }
  ],
  "connections": [
    { "from": "토픽1", "to": "토픽2", "reason": "연결 이유" }
  ]
}
```

### 비용 최적화

| 전략 | 설명 |
|-----|------|
| **URL 중복 제거** | 같은 URL은 태그 1번만 추출 (url_hash 기반) |
| **배치 처리** | 이벤트 수신 시 여러 URL 한 번에 처리 |
| **경량 모델 사용** | 태그 추출은 GPT-3.5/Gemini Flash로 충분 |
| **관계도만 고급 모델** | 세션당 1회이므로 GPT-4/Claude 사용 가능 |

---

## 아키텍처 개요

```mermaid
flowchart TB
    subgraph Service_Layer
        TS[TagExtractionService]
        MS[MindmapService]
    end

    subgraph AI_Provider_Layer
        IF[AIProvider Interface]
        IF --> OAI[OpenAI Provider]
        IF --> GEM[Google Gemini Provider]
        IF --> CLA[Anthropic Claude Provider]
    end

    subgraph Infrastructure
        PM[ProviderManager]
        CFG[Config]
    end

    TS --> PM
    MS --> PM
    PM --> IF
    CFG --> PM
```

### 설계 원칙

1. **인터페이스 기반 설계**: 모든 AI 프로바이더는 동일한 인터페이스 구현
2. **런타임 프로바이더 전환**: 환경변수 또는 설정으로 프로바이더 변경 가능
3. **Fallback 지원**: 기본 프로바이더 실패 시 대체 프로바이더 사용
4. **용도별 프로바이더 분리**: 태그 추출과 관계도 생성에 다른 모델 사용 가능

---

## 진행 상황

| Step | 이름 | 상태 |
|------|------|------|
| 9.1 | AI Provider 인터페이스 정의 | ⬜ |
| 9.2 | 개별 Provider 구현 (OpenAI, Gemini, Claude) | ⬜ |
| 9.3 | Provider Manager 및 Config | ⬜ |
| 9.4 | 태그 추출 서비스 (페이지당) | ⬜ |
| 9.5 | 마인드맵 생성 서비스 (세션당) | ⬜ |

---

## Step 9.1: AI Provider 인터페이스 정의

### 체크리스트

- [ ] **공통 타입 정의**
  - [ ] `internal/infrastructure/ai/types.go`
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
  - [ ] `internal/infrastructure/ai/provider.go`
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
go build ./...
# 컴파일 성공
```

---

## Step 9.2: 개별 Provider 구현

### 체크리스트

- [ ] **의존성 추가**
  ```bash
  # OpenAI
  go get github.com/sashabaranov/go-openai

  # Google Gemini
  go get github.com/google/generative-ai-go

  # Anthropic Claude
  go get github.com/anthropics/anthropic-sdk-go
  ```

- [ ] **OpenAI Provider 구현**
  - [ ] `internal/infrastructure/ai/openai.go`
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
        // Simple health check with minimal tokens
        _, err := p.Chat(ctx, []Message{
            {Role: RoleUser, Content: "ping"},
        }, ChatOptions{MaxTokens: 5})
        return err == nil
    }
    ```

- [ ] **Google Gemini Provider 구현**
  - [ ] `internal/infrastructure/ai/gemini.go`
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

        // Set generation config
        model.SetTemperature(float32(opts.Temperature))
        model.SetMaxOutputTokens(int32(opts.MaxTokens))
        model.SetTopP(float32(opts.TopP))

        if len(opts.StopSequences) > 0 {
            model.StopSequences = opts.StopSequences
        }

        // Convert messages to Gemini format
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

        // Extract text from response
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

        // Set generation config
        model.SetTemperature(float32(opts.Temperature))
        model.SetMaxOutputTokens(int32(opts.MaxTokens))
        model.SetTopP(float32(opts.TopP))

        // Force JSON output
        model.ResponseMIMEType = "application/json"

        // Convert messages
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

        // Validate JSON
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
  - [ ] `internal/infrastructure/ai/claude.go`
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
        // Separate system message from conversation
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

        // Extract text content
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
        // Add JSON instruction to the last user message
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

        // Validate JSON
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
go build ./...
# 컴파일 성공
```

---

## Step 9.3: Provider Manager 및 Config

### 체크리스트

- [ ] **환경 변수 설정**
  ```env
  # AI Provider 설정
  AI_DEFAULT_PROVIDER=openai
  AI_FALLBACK_PROVIDERS=gemini,claude

  # 용도별 프로바이더 (선택적)
  AI_SUMMARIZE_PROVIDER=gemini
  AI_MINDMAP_PROVIDER=openai

  # OpenAI
  OPENAI_API_KEY=sk-...
  OPENAI_MODEL=gpt-4-turbo-preview

  # Google Gemini
  GEMINI_API_KEY=...
  GEMINI_MODEL=gemini-1.5-pro

  # Anthropic Claude
  CLAUDE_API_KEY=sk-ant-...
  CLAUDE_MODEL=claude-3-5-sonnet-20241022
  ```

- [ ] **Config 업데이트**
  - [ ] `internal/infrastructure/config/config.go`
    ```go
    type Config struct {
        // ... 기존 필드

        // AI Settings
        AI AIConfig
    }

    type AIConfig struct {
        // Default provider for general use
        DefaultProvider   string   `json:"default_provider"`
        FallbackProviders []string `json:"fallback_providers"`

        // Task-specific providers (optional override)
        SummarizeProvider string `json:"summarize_provider"`
        MindmapProvider   string `json:"mindmap_provider"`

        // Provider configurations
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

    func Load() *Config {
        return &Config{
            // ... 기존 필드

            AI: AIConfig{
                DefaultProvider:   getEnv("AI_DEFAULT_PROVIDER", "openai"),
                FallbackProviders: getEnvSlice("AI_FALLBACK_PROVIDERS", []string{}),
                SummarizeProvider: getEnv("AI_SUMMARIZE_PROVIDER", ""),
                MindmapProvider:   getEnv("AI_MINDMAP_PROVIDER", ""),

                OpenAI: OpenAIConfig{
                    APIKey:  getEnv("OPENAI_API_KEY", ""),
                    Model:   getEnv("OPENAI_MODEL", "gpt-4-turbo-preview"),
                    Enabled: getEnv("OPENAI_API_KEY", "") != "",
                },
                Gemini: GeminiConfig{
                    APIKey:  getEnv("GEMINI_API_KEY", ""),
                    Model:   getEnv("GEMINI_MODEL", "gemini-1.5-pro"),
                    Enabled: getEnv("GEMINI_API_KEY", "") != "",
                },
                Claude: ClaudeConfig{
                    APIKey:  getEnv("CLAUDE_API_KEY", ""),
                    Model:   getEnv("CLAUDE_MODEL", "claude-3-5-sonnet-20241022"),
                    Enabled: getEnv("CLAUDE_API_KEY", "") != "",
                },
            },
        }
    }

    func getEnvSlice(key string, defaultVal []string) []string {
        val := os.Getenv(key)
        if val == "" {
            return defaultVal
        }
        return strings.Split(val, ",")
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

        "github.com/mindhit/api/internal/infrastructure/config"
    )

    // TaskType identifies the AI task for provider selection
    type TaskType string

    const (
        TaskSummarize TaskType = "summarize"
        TaskMindmap   TaskType = "mindmap"
        TaskGeneral   TaskType = "general"
    )

    // ProviderManager manages multiple AI providers with fallback support
    type ProviderManager struct {
        providers         map[ProviderType]AIProvider
        defaultProvider   ProviderType
        fallbackOrder     []ProviderType
        taskProviders     map[TaskType]ProviderType
        mu                sync.RWMutex
    }

    // NewProviderManager creates a new provider manager from config
    func NewProviderManager(ctx context.Context, cfg config.AIConfig) (*ProviderManager, error) {
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
            // Fall back to first available provider
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
        if cfg.SummarizeProvider != "" {
            pt := ProviderType(cfg.SummarizeProvider)
            if _, ok := pm.providers[pt]; ok {
                pm.taskProviders[TaskSummarize] = pt
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

    // GetProvider returns the provider for a specific task
    func (pm *ProviderManager) GetProvider(task TaskType) AIProvider {
        pm.mu.RLock()
        defer pm.mu.RUnlock()

        // Check task-specific provider first
        if pt, ok := pm.taskProviders[task]; ok {
            if provider, ok := pm.providers[pt]; ok {
                return provider
            }
        }

        // Return default provider
        return pm.providers[pm.defaultProvider]
    }

    // GetProviderByType returns a specific provider by type
    func (pm *ProviderManager) GetProviderByType(pt ProviderType) (AIProvider, bool) {
        pm.mu.RLock()
        defer pm.mu.RUnlock()

        provider, ok := pm.providers[pt]
        return provider, ok
    }

    // Chat executes chat with automatic fallback
    func (pm *ProviderManager) Chat(ctx context.Context, task TaskType, messages []Message, opts ChatOptions) (*ChatResponse, error) {
        return pm.chatWithFallback(ctx, task, messages, opts, false)
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

        // Task-specific provider first
        if pt, ok := pm.taskProviders[task]; ok {
            if p, ok := pm.providers[pt]; ok {
                result = append(result, p)
                seen[pt] = true
            }
        }

        // Default provider
        if !seen[pm.defaultProvider] {
            if p, ok := pm.providers[pm.defaultProvider]; ok {
                result = append(result, p)
                seen[pm.defaultProvider] = true
            }
        }

        // Fallback providers
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

    // HealthCheck checks all providers
    func (pm *ProviderManager) HealthCheck(ctx context.Context) map[ProviderType]bool {
        pm.mu.RLock()
        defer pm.mu.RUnlock()

        results := make(map[ProviderType]bool)
        for pt, provider := range pm.providers {
            results[pt] = provider.IsHealthy(ctx)
        }
        return results
    }

    // Close closes all providers that implement io.Closer
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

- [ ] **main.go에 Provider Manager 초기화**
  ```go
  import "github.com/mindhit/api/internal/infrastructure/ai"

  // Initialize AI Provider Manager
  aiManager, err := ai.NewProviderManager(ctx, cfg.AI)
  if err != nil {
      slog.Error("failed to initialize ai manager", "error", err)
      os.Exit(1)
  }
  defer aiManager.Close()
  ```

### 검증
```bash
go build ./...
# 컴파일 성공
```

---

## Step 9.4: 태그 추출 서비스 (페이지당)

### 목표

이벤트 배치 수신 시 새로운 URL에 대해 실시간으로 태그를 추출합니다.

### 체크리스트

- [ ] **태그 추출 서비스**
  - [ ] `internal/service/tag_extraction_service.go`
    ```go
    package service

    import (
        "context"
        "encoding/json"
        "fmt"
        "log/slog"

        "github.com/google/uuid"
        "github.com/mindhit/api/ent"
        "github.com/mindhit/api/ent/url"
        "github.com/mindhit/api/internal/infrastructure/ai"
    )

    type TagExtractionService struct {
        client    *ent.Client
        aiManager *ai.ProviderManager
    }

    func NewTagExtractionService(client *ent.Client, aiManager *ai.ProviderManager) *TagExtractionService {
        return &TagExtractionService{
            client:    client,
            aiManager: aiManager,
        }
    }

    type TagResult struct {
        Tags    []string `json:"tags"`
        Summary string   `json:"summary"`
    }

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

    // ExtractTags extracts tags from a URL's content
    // Called when a new URL is received in event batch
    func (s *TagExtractionService) ExtractTags(ctx context.Context, urlID uuid.UUID) error {
        // Get URL from database
        u, err := s.client.URL.Get(ctx, urlID)
        if err != nil {
            return fmt.Errorf("get url: %w", err)
        }

        // Skip if already has tags (중복 URL 처리)
        if len(u.Tags) > 0 {
            slog.Debug("url already has tags, skipping", "url", u.URL)
            return nil
        }

        // Skip if no content (Extension에서 추출 실패한 경우)
        if u.Content == "" {
            slog.Warn("url has no content, skipping tag extraction", "url", u.URL)
            return nil
        }

        // Generate tags using AI (경량 모델 사용)
        messages := []ai.Message{
            {
                Role:    ai.RoleUser,
                Content: fmt.Sprintf(tagExtractionPrompt, u.Title, truncateContent(u.Content, 8000)),
            },
        }

        opts := ai.DefaultChatOptions()
        opts.MaxTokens = 500 // 태그 추출은 짧은 응답

        response, err := s.aiManager.ChatWithJSON(ctx, ai.TaskTagExtraction, messages, opts)
        if err != nil {
            return fmt.Errorf("ai tag extraction: %w", err)
        }

        var result TagResult
        if err := json.Unmarshal([]byte(response.Content), &result); err != nil {
            return fmt.Errorf("parse ai response: %w", err)
        }

        // Update URL with tags and summary
        _, err = s.client.URL.UpdateOneID(urlID).
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

    // ExtractTagsForNewURLs processes multiple new URLs from an event batch
    func (s *TagExtractionService) ExtractTagsForNewURLs(ctx context.Context, urlIDs []uuid.UUID) error {
        for _, urlID := range urlIDs {
            if err := s.ExtractTags(ctx, urlID); err != nil {
                slog.Error("failed to extract tags",
                    "url_id", urlID,
                    "error", err,
                )
                // Continue with other URLs (graceful degradation)
            }
        }
        return nil
    }

    func truncateContent(content string, maxLen int) string {
        if len(content) <= maxLen {
            return content
        }
        return content[:maxLen] + "..."
    }
    ```

- [ ] **이벤트 서비스에 태그 추출 통합**
  - [ ] `internal/service/event_service.go` 수정
    ```go
    type EventService struct {
        client               *ent.Client
        tagExtractionService *TagExtractionService
    }

    func (s *EventService) ProcessBatchEvents(ctx context.Context, sessionID uuid.UUID, events []Event) error {
        var newURLIDs []uuid.UUID

        for _, event := range events {
            if event.Type == "page_visit" {
                // URL 저장 또는 조회
                urlID, isNew, err := s.upsertURL(ctx, event)
                if err != nil {
                    slog.Error("failed to upsert url", "error", err)
                    continue
                }

                // 새 URL인 경우 태그 추출 대상에 추가
                if isNew {
                    newURLIDs = append(newURLIDs, urlID)
                }

                // page_visit 저장
                // ...
            }
        }

        // 새 URL들에 대해 비동기로 태그 추출
        if len(newURLIDs) > 0 {
            go func() {
                bgCtx := context.Background()
                if err := s.tagExtractionService.ExtractTagsForNewURLs(bgCtx, newURLIDs); err != nil {
                    slog.Error("failed to extract tags for new urls", "error", err)
                }
            }()
        }

        return nil
    }

    func (s *EventService) upsertURL(ctx context.Context, event Event) (uuid.UUID, bool, error) {
        urlHash := hashURL(event.URL)

        // 기존 URL 조회
        existing, err := s.client.URL.Query().
            Where(url.URLHashEQ(urlHash)).
            Only(ctx)

        if err == nil {
            // 이미 존재하는 URL
            return existing.ID, false, nil
        }

        if !ent.IsNotFound(err) {
            return uuid.Nil, false, err
        }

        // 새 URL 생성
        newURL, err := s.client.URL.Create().
            SetURL(event.URL).
            SetURLHash(urlHash).
            SetTitle(event.Title).
            SetContent(event.Content). // Extension에서 추출한 콘텐츠
            Save(ctx)

        if err != nil {
            return uuid.Nil, false, err
        }

        return newURL.ID, true, nil
    }
    ```

- [ ] **AI Config에 TaskTagExtraction 추가**
  ```go
  // internal/infrastructure/ai/manager.go
  const (
      TaskSummarize      TaskType = "summarize"
      TaskMindmap        TaskType = "mindmap"
      TaskTagExtraction  TaskType = "tag_extraction"  // NEW
      TaskGeneral        TaskType = "general"
  )
  ```

- [ ] **환경변수에 태그 추출 프로바이더 설정**
  ```env
  # 태그 추출은 경량 모델 사용 (비용 최적화)
  AI_TAG_EXTRACTION_PROVIDER=gemini
  GEMINI_MODEL=gemini-1.5-flash  # 빠르고 저렴한 모델
  ```

### 검증

```bash
go build ./...
# 컴파일 성공

# 테스트 (유닛 테스트)
go test ./internal/service/...

# 로그에서 태그 추출 확인
# "extracted tags" url=https://... tags=["AI", "머신러닝"] provider=gemini
```

---

## Step 9.5: 관계도 생성 서비스 (세션 종료 시 1회)

### 목표

세션 종료 시 추출된 태그들을 기반으로 LLM이 관계도 JSON을 생성합니다.
- **입력**: 세션 내 모든 URL + 각 URL의 태그/요약
- **출력**: 마인드맵 구조 JSON (nodes, edges, layout)
- **호출 시점**: 세션 Stop 시 1회만 호출

### 체크리스트

- [ ] **마인드맵 타입 정의**
  - [ ] `internal/service/mindmap_types.go`
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
        Label  string  `json:"label,omitempty"` // 연결 이유
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

- [ ] **관계도 생성 서비스** (태그 기반)
  - [ ] `internal/service/mindmap_service.go`
    ```go
    package service

    import (
        "context"
        "encoding/json"
        "fmt"
        "log/slog"
        "math"
        "strings"

        "github.com/google/uuid"
        "github.com/mindhit/api/ent"
        "github.com/mindhit/api/ent/session"
        "github.com/mindhit/api/internal/infrastructure/ai"
    )

    type MindmapService struct {
        client    *ent.Client
        aiManager *ai.ProviderManager
    }

    func NewMindmapService(client *ent.Client, aiManager *ai.ProviderManager) *MindmapService {
        return &MindmapService{
            client:    client,
            aiManager: aiManager,
        }
    }

    // 태그 기반 관계도 생성 프롬프트
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

    // AI 응답 타입 (태그 기반)
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

    // GenerateRelationshipGraph generates a mindmap from extracted tags
    // Called once when session is stopped
    func (s *MindmapService) GenerateRelationshipGraph(ctx context.Context, sessionID uuid.UUID) error {
        // Get session with all related data (URLs already have tags from Step 9.4)
        sess, err := s.client.Session.
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

        // Build page data with tags (Step 9.4에서 추출된 태그 사용)
        var pageData strings.Builder
        urlMap := make(map[string]*ent.URL)
        dwellTimeMap := make(map[string]int)

        for _, pv := range sess.Edges.PageVisits {
            if pv.Edges.URL == nil {
                continue
            }
            u := pv.Edges.URL
            urlMap[u.ID.String()] = u

            dwellTime := 0
            if pv.DwellTimeSeconds != nil {
                dwellTime = *pv.DwellTimeSeconds
            }
            dwellTimeMap[u.ID.String()] = dwellTime

            // 태그와 요약 포함 (Step 9.4에서 이미 추출됨)
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

        // Generate relationship graph using AI (프리미엄 모델 사용)
        messages := []ai.Message{
            {
                Role:    ai.RoleUser,
                Content: fmt.Sprintf(relationshipGraphPrompt, pageData.String(), highlights.String()),
            },
        }

        opts := ai.DefaultChatOptions()
        opts.MaxTokens = 4096 // 관계도 JSON은 적당한 크기

        response, err := s.aiManager.ChatWithJSON(ctx, ai.TaskMindmap, messages, opts)
        if err != nil {
            return fmt.Errorf("ai generate relationship graph: %w", err)
        }

        var aiResp RelationshipGraphResponse
        if err := json.Unmarshal([]byte(response.Content), &aiResp); err != nil {
            return fmt.Errorf("parse ai response: %w", err)
        }

        // Convert AI response to mindmap data
        mindmapData := s.buildMindmapFromRelationship(aiResp, urlMap, dwellTimeMap)

        // Save mindmap to database
        _, err = s.client.MindmapGraph.
            Create().
            SetSessionID(sessionID).
            SetNodes(mindmapData.Nodes).
            SetEdges(mindmapData.Edges).
            SetLayout(mindmapData.Layout).
            Save(ctx)

        if err != nil {
            return fmt.Errorf("save mindmap: %w", err)
        }

        // Update session status
        _, err = s.client.Session.
            UpdateOneID(sessionID).
            SetSessionStatus(session.SessionStatusCompleted).
            Save(ctx)

        if err != nil {
            return fmt.Errorf("update session status: %w", err)
        }

        slog.Info("generated relationship graph",
            "session_id", sessionID,
            "topics", len(aiResp.Topics),
            "connections", len(aiResp.Connections),
            "provider", response.Provider,
            "tokens", response.InputTokens+response.OutputTokens,
        )
        return nil
    }

    // buildMindmapFromRelationship converts AI response to mindmap data structure
    func (s *MindmapService) buildMindmapFromRelationship(resp RelationshipGraphResponse, urlMap map[string]*ent.URL, dwellTimeMap map[string]int) MindmapData {
        var nodes []MindmapNode
        var edges []MindmapEdge

        // Create core node (center of galaxy - 태양)
        coreID := "core"
        nodes = append(nodes, MindmapNode{
            ID:    coreID,
            Label: resp.Core.Label,
            Type:  "core",
            Size:  100,
            Color: "#FFD700", // Gold for core
            Position: &Position{X: 0, Y: 0, Z: 0},
            Data: map[string]interface{}{
                "description": resp.Core.Description,
            },
        })

        // Create topic nodes (planets orbiting the sun - 행성)
        topicCount := len(resp.Topics)
        for i, topic := range resp.Topics {
            topicID := topic.ID
            if topicID == "" {
                topicID = fmt.Sprintf("topic-%d", i)
            }

            // Position in orbit around core
            angle := (float64(i) / float64(topicCount)) * 2 * math.Pi
            radius := 200.0

            // 토픽 크기 = 연결된 페이지 수에 비례
            topicSize := 40.0 + float64(len(topic.Pages))*10
            if topicSize > 80 {
                topicSize = 80
            }

            nodes = append(nodes, MindmapNode{
                ID:    topicID,
                Label: topic.Label,
                Type:  "topic",
                Size:  topicSize,
                Color: getTopicColor(i),
                Position: &Position{
                    X: radius * math.Cos(angle),
                    Y: radius * math.Sin(angle),
                    Z: 0,
                },
                Data: map[string]interface{}{
                    "description": topic.Description,
                    "tags":        topic.Tags,
                },
            })

            // Edge from core to topic
            edges = append(edges, MindmapEdge{
                Source: coreID,
                Target: topicID,
                Weight: 1.0,
            })

            // Create page nodes (moons - 위성)
            for j, page := range topic.Pages {
                pageID := page.URLID

                // Position around parent topic
                subAngle := angle + (float64(j)-float64(len(topic.Pages))/2)*0.4
                subRadius := 60.0 + float64(j)*15

                // Calculate size based on dwell time
                size := 15.0
                if dwell, ok := dwellTimeMap[page.URLID]; ok {
                    size = math.Min(40, 15+float64(dwell)/20)
                }

                // relevance가 높을수록 더 큰 크기
                size = size * (0.5 + page.Relevance*0.5)

                nodes = append(nodes, MindmapNode{
                    ID:    pageID,
                    Label: page.Title,
                    Type:  "page",
                    Size:  size,
                    Color: getTopicColor(i),
                    Position: &Position{
                        X: radius*math.Cos(angle) + subRadius*math.Cos(subAngle),
                        Y: radius*math.Sin(angle) + subRadius*math.Sin(subAngle),
                        Z: 0,
                    },
                    Data: map[string]interface{}{
                        "url_id":    page.URLID,
                        "relevance": page.Relevance,
                    },
                })

                edges = append(edges, MindmapEdge{
                    Source: topicID,
                    Target: pageID,
                    Weight: page.Relevance,
                })
            }
        }

        // Add cross-topic connections (공통 태그 기반 연결)
        for _, conn := range resp.Connections {
            edges = append(edges, MindmapEdge{
                Source: conn.From,
                Target: conn.To,
                Weight: float64(len(conn.SharedTags)) * 0.2, // 공통 태그 수에 비례
                Label:  conn.Reason,
            })
        }

        return MindmapData{
            Nodes: nodes,
            Edges: edges,
            Layout: MindmapLayout{
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
            "#3B82F6", // Blue
            "#10B981", // Green
            "#F59E0B", // Amber
            "#EF4444", // Red
            "#8B5CF6", // Purple
            "#EC4899", // Pink
            "#14B8A6", // Teal
            "#F97316", // Orange
        }
        return colors[index%len(colors)]
    }
    ```

- [ ] **관계도 생성 Job** (세션 종료 시 실행)
  - [ ] `internal/jobs/relationship_graph_job.go`
    ```go
    package jobs

    import (
        "context"
        "log/slog"
        "time"

        "github.com/google/uuid"
        "github.com/mindhit/api/ent"
        "github.com/mindhit/api/ent/session"
        "github.com/mindhit/api/internal/service"
    )

    // RelationshipGraphJob processes stopped sessions to generate relationship graphs
    // 태그 추출은 이미 Step 9.4에서 실시간으로 처리됨
    // 이 Job은 관계도 생성만 담당
    type RelationshipGraphJob struct {
        client         *ent.Client
        mindmapService *service.MindmapService
    }

    func NewRelationshipGraphJob(
        client *ent.Client,
        mindmapService *service.MindmapService,
    ) *RelationshipGraphJob {
        return &RelationshipGraphJob{
            client:         client,
            mindmapService: mindmapService,
        }
    }

    func (j *RelationshipGraphJob) Run() {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
        defer cancel()

        // Find sessions in "processing" status (세션이 Stop되면 processing 상태)
        sessions, err := j.client.Session.
            Query().
            Where(session.SessionStatusEQ(session.SessionStatusProcessing)).
            Limit(5). // Process 5 at a time
            All(ctx)

        if err != nil {
            slog.Error("failed to get processing sessions", "error", err)
            return
        }

        for _, sess := range sessions {
            slog.Info("generating relationship graph", "session_id", sess.ID)

            // 관계도 생성 (태그는 이미 추출되어 있음)
            if err := j.mindmapService.GenerateRelationshipGraph(ctx, sess.ID); err != nil {
                slog.Error("failed to generate relationship graph",
                    "session_id", sess.ID,
                    "error", err,
                )
                j.markSessionFailed(ctx, sess.ID)
                continue
            }

            slog.Info("relationship graph generated", "session_id", sess.ID)
        }
    }

    func (j *RelationshipGraphJob) markSessionFailed(ctx context.Context, sessionID uuid.UUID) {
        _, err := j.client.Session.
            UpdateOneID(sessionID).
            SetSessionStatus(session.SessionStatusFailed).
            Save(ctx)

        if err != nil {
            slog.Error("failed to mark session as failed",
                "session_id", sessionID,
                "error", err,
            )
        }
    }
    ```

- [ ] **main.go에 서비스 및 Job 등록**
  ```go
  import (
      "github.com/mindhit/api/internal/infrastructure/ai"
      "github.com/mindhit/api/internal/jobs"
      "github.com/mindhit/api/internal/service"
  )

  // Initialize AI Provider Manager
  aiManager, err := ai.NewProviderManager(ctx, cfg.AI)
  if err != nil {
      slog.Error("failed to initialize ai manager", "error", err)
      os.Exit(1)
  }
  defer aiManager.Close()

  // Initialize services
  tagExtractionService := service.NewTagExtractionService(client, aiManager)
  mindmapService := service.NewMindmapService(client, aiManager)
  eventService := service.NewEventService(client, tagExtractionService)

  // Register relationship graph job (every 2 minutes)
  // 세션 Stop 후 관계도 생성
  relationshipGraphJob := jobs.NewRelationshipGraphJob(client, mindmapService)
  if err := sched.RegisterIntervalJob("relationship-graph", 2*time.Minute, relationshipGraphJob.Run); err != nil {
      slog.Error("failed to register relationship graph job", "error", err)
  }
  ```

- [ ] **관계도 API 엔드포인트**
  - [ ] `internal/controller/mindmap_controller.go`
    ```go
    package controller

    import (
        "context"
        "log/slog"
        "net/http"

        "github.com/gin-gonic/gin"
        "github.com/google/uuid"
        "github.com/mindhit/api/ent"
        "github.com/mindhit/api/ent/mindmapgraph"
        "github.com/mindhit/api/ent/session"
        "github.com/mindhit/api/ent/user"
        "github.com/mindhit/api/internal/infrastructure/middleware"
        "github.com/mindhit/api/internal/service"
    )

    type MindmapController struct {
        client         *ent.Client
        mindmapService *service.MindmapService
    }

    func NewMindmapController(
        client *ent.Client,
        mindmapService *service.MindmapService,
    ) *MindmapController {
        return &MindmapController{
            client:         client,
            mindmapService: mindmapService,
        }
    }

    // GetBySession retrieves the relationship graph for a session
    func (c *MindmapController) GetBySession(ctx *gin.Context) {
        sessionID, err := uuid.Parse(ctx.Param("id"))
        if err != nil {
            ctx.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid session id"}})
            return
        }

        mindmap, err := c.client.MindmapGraph.
            Query().
            Where(mindmapgraph.HasSessionWith(session.IDEQ(sessionID))).
            Only(ctx.Request.Context())

        if err != nil {
            if ent.IsNotFound(err) {
                ctx.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "mindmap not found"}})
                return
            }
            ctx.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
            return
        }

        ctx.JSON(http.StatusOK, gin.H{"mindmap": mindmap})
    }

    // Generate triggers relationship graph generation for a session
    // 태그는 이미 Step 9.4에서 추출되어 있음
    func (c *MindmapController) Generate(ctx *gin.Context) {
        userID, ok := middleware.GetUserID(ctx)
        if !ok {
            ctx.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"message": "unauthorized"}})
            return
        }

        sessionID, err := uuid.Parse(ctx.Param("id"))
        if err != nil {
            ctx.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid session id"}})
            return
        }

        // Verify session ownership and status
        sess, err := c.client.Session.
            Query().
            Where(
                session.IDEQ(sessionID),
                session.HasUserWith(user.IDEQ(userID)),
            ).
            Only(ctx.Request.Context())

        if err != nil {
            if ent.IsNotFound(err) {
                ctx.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "session not found"}})
                return
            }
            ctx.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
            return
        }

        // Check if session is in valid state for mindmap generation
        if sess.SessionStatus != session.SessionStatusProcessing && sess.SessionStatus != session.SessionStatusCompleted {
            ctx.JSON(http.StatusBadRequest, gin.H{
                "error": gin.H{"message": "session must be stopped before generating mindmap"},
            })
            return
        }

        // Check if mindmap already exists
        exists, err := c.client.MindmapGraph.
            Query().
            Where(mindmapgraph.HasSessionWith(session.IDEQ(sessionID))).
            Exist(ctx.Request.Context())

        if err != nil {
            ctx.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
            return
        }

        if exists {
            ctx.JSON(http.StatusConflict, gin.H{
                "error": gin.H{"message": "mindmap already exists for this session"},
            })
            return
        }

        // Start async processing
        // 태그는 이미 추출되어 있으므로 관계도만 생성
        go func() {
            bgCtx := context.Background()
            if err := c.mindmapService.GenerateRelationshipGraph(bgCtx, sessionID); err != nil {
                slog.Error("failed to generate relationship graph", "session_id", sessionID, "error", err)
                return
            }
        }()

        ctx.JSON(http.StatusAccepted, gin.H{
            "message":    "relationship graph generation started",
            "session_id": sessionID.String(),
        })
    }

    // Regenerate regenerates the mindmap for a session (replaces existing)
    func (c *MindmapController) Regenerate(ctx *gin.Context) {
        userID, ok := middleware.GetUserID(ctx)
        if !ok {
            ctx.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"message": "unauthorized"}})
            return
        }

        sessionID, err := uuid.Parse(ctx.Param("id"))
        if err != nil {
            ctx.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid session id"}})
            return
        }

        // Verify session ownership
        _, err = c.client.Session.
            Query().
            Where(
                session.IDEQ(sessionID),
                session.HasUserWith(user.IDEQ(userID)),
            ).
            Only(ctx.Request.Context())

        if err != nil {
            if ent.IsNotFound(err) {
                ctx.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "session not found"}})
                return
            }
            ctx.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
            return
        }

        // Delete existing mindmap if exists
        _, err = c.client.MindmapGraph.
            Delete().
            Where(mindmapgraph.HasSessionWith(session.IDEQ(sessionID))).
            Exec(ctx.Request.Context())

        if err != nil && !ent.IsNotFound(err) {
            ctx.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
            return
        }

        // Start async processing
        go func() {
            bgCtx := context.Background()

            if err := c.mindmapService.GenerateMindmap(bgCtx, sessionID); err != nil {
                slog.Error("failed to regenerate mindmap", "session_id", sessionID, "error", err)
                return
            }
        }()

        ctx.JSON(http.StatusAccepted, gin.H{
            "message":    "mindmap regeneration started",
            "session_id": sessionID.String(),
        })
    }
    ```

- [ ] **라우터에 마인드맵 엔드포인트 추가**
  ```go
  // In main.go or router setup
  mindmapController := controller.NewMindmapController(client, summarizeService, mindmapService)

  sessions := v1.Group("/sessions")
  sessions.Use(middleware.Auth(jwtService))
  {
      // ... 기존 세션 라우트

      // Mindmap
      sessions.GET("/:id/mindmap", mindmapController.GetBySession)
      sessions.POST("/:id/mindmap/generate", mindmapController.Generate)
      sessions.POST("/:id/mindmap/regenerate", mindmapController.Regenerate)
  }
  ```

- [ ] **AI 헬스체크 엔드포인트** (선택)
  ```go
  // Health check for AI providers
  r.GET("/health/ai", func(ctx *gin.Context) {
      health := aiManager.HealthCheck(ctx.Request.Context())
      allHealthy := true
      for _, ok := range health {
          if !ok {
              allHealthy = false
              break
          }
      }

      status := http.StatusOK
      if !allHealthy {
          status = http.StatusServiceUnavailable
      }

      ctx.JSON(status, gin.H{"providers": health})
  })
  ```

### 검증
```bash
# 서버 실행
go run ./cmd/server

# 세션 종료 후 5분 이내 마인드맵 생성 확인
# 로그에서 provider와 model 정보 확인
# "generated mindmap" session_id=... provider=openai model=gpt-4-turbo

# AI 헬스체크
curl http://localhost:8080/health/ai
# {"providers":{"claude":true,"gemini":true,"openai":true}}

# API로 마인드맵 조회
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/v1/sessions/{session_id}/mindmap
```

---

## 프로바이더 비교

| 프로바이더 | 모델 | 장점 | 단점 | 권장 용도 |
|-----------|------|------|------|----------|
| **OpenAI** | gpt-4-turbo | JSON mode 네이티브 지원, 안정적 | 비용 높음 | 마인드맵 생성 |
| **Gemini** | gemini-1.5-pro | 긴 컨텍스트, 비용 효율적 | JSON 출력 불안정할 수 있음 | URL 요약 |
| **Claude** | claude-3-5-sonnet | 뛰어난 분석력, 한국어 우수 | JSON mode 없음 | 복잡한 분석 |

### 권장 설정

```env
# 비용 최적화 설정
AI_DEFAULT_PROVIDER=gemini
AI_FALLBACK_PROVIDERS=openai,claude
AI_SUMMARIZE_PROVIDER=gemini      # 저비용
AI_MINDMAP_PROVIDER=openai        # JSON 안정성

# 품질 최적화 설정
AI_DEFAULT_PROVIDER=openai
AI_FALLBACK_PROVIDERS=claude,gemini
AI_SUMMARIZE_PROVIDER=claude      # 분석력
AI_MINDMAP_PROVIDER=openai        # JSON mode
```

---

## API 요약

| Method | Endpoint | 설명 | 인증 |
|--------|----------|------|------|
| GET | `/v1/sessions/:id/mindmap` | 세션의 마인드맵 조회 | Bearer Token |
| POST | `/v1/sessions/:id/mindmap/generate` | 마인드맵 생성 시작 (비동기) | Bearer Token |
| POST | `/v1/sessions/:id/mindmap/regenerate` | 마인드맵 재생성 (기존 삭제 후 새로 생성) | Bearer Token |
| GET | `/health/ai` | AI 프로바이더 헬스체크 | - |

### 응답 예시

**GET /v1/sessions/:id/mindmap** (성공)
```json
{
  "mindmap": {
    "id": "uuid",
    "session_id": "uuid",
    "nodes": [
      {
        "id": "node-1",
        "label": "핵심 주제",
        "type": "core",
        "size": 100,
        "color": "#FFD700",
        "position": {"x": 0, "y": 0, "z": 0}
      }
    ],
    "edges": [
      {"source": "node-1", "target": "node-2", "weight": 1.0}
    ],
    "layout": {
      "type": "galaxy",
      "params": {"center": [0, 0, 0], "scale": 1.0}
    },
    "generated_at": "2024-01-01T00:00:00Z"
  }
}
```

**GET /health/ai** (성공)
```json
{
  "providers": {
    "openai": true,
    "gemini": true,
    "claude": false
  }
}
```

---

## Phase 9 완료 확인

### 전체 검증 체크리스트

- [ ] AI Provider 인터페이스 정의
- [ ] OpenAI Provider 구현
- [ ] Google Gemini Provider 구현
- [ ] Anthropic Claude Provider 구현
- [ ] Provider Manager 구현 (Fallback 지원)
- [ ] URL 콘텐츠 크롤링
- [ ] URL 요약 생성 (다중 프로바이더)
- [ ] 마인드맵 구조 생성 (다중 프로바이더)
- [ ] 마인드맵 저장
- [ ] 마인드맵 API 조회
- [ ] 5분마다 자동 처리
- [ ] AI 헬스체크 엔드포인트

### 산출물 요약

| 항목 | 위치 |
|-----|------|
| AI 타입 정의 | `internal/infrastructure/ai/types.go` |
| Provider 인터페이스 | `internal/infrastructure/ai/provider.go` |
| OpenAI Provider | `internal/infrastructure/ai/openai.go` |
| Gemini Provider | `internal/infrastructure/ai/gemini.go` |
| Claude Provider | `internal/infrastructure/ai/claude.go` |
| Provider Manager | `internal/infrastructure/ai/manager.go` |
| 크롤러 | `internal/infrastructure/crawler/crawler.go` |
| 요약 서비스 | `internal/service/summarize_service.go` |
| 마인드맵 서비스 | `internal/service/mindmap_service.go` |
| AI 처리 Job | `internal/jobs/ai_processing.go` |
| 마인드맵 API | `internal/controller/mindmap_controller.go` |

---

## 다음 Phase

Phase 9 완료 후 [Phase 10: 웹앱 대시보드](./phase-10-dashboard.md)으로 진행하세요.
