// Package ai provides unified AI provider integration for multiple providers.
package ai

import "time"

// ProviderType identifies the AI provider.
type ProviderType string

const (
	// ProviderOpenAI represents the OpenAI provider.
	ProviderOpenAI ProviderType = "openai"
	// ProviderGemini represents the Google Gemini provider.
	ProviderGemini ProviderType = "gemini"
	// ProviderClaude represents the Anthropic Claude provider.
	ProviderClaude ProviderType = "claude"
)

// TaskType identifies the AI task for provider selection.
type TaskType string

const (
	// TaskTagExtraction is used for extracting tags from page content.
	TaskTagExtraction TaskType = "tag_extraction"
	// TaskMindmap is used for generating mindmaps from sessions.
	TaskMindmap TaskType = "mindmap"
	// TaskGeneral is used for general AI tasks.
	TaskGeneral TaskType = "general"
)

// Role defines message roles in a conversation.
type Role string

const (
	// RoleSystem is the system message role.
	RoleSystem Role = "system"
	// RoleUser is the user message role.
	RoleUser Role = "user"
	// RoleAssistant is the assistant message role.
	RoleAssistant Role = "assistant"
)

// Message represents a chat message in a conversation.
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// ChatOptions contains optional parameters for chat completion.
type ChatOptions struct {
	Temperature    float64  `json:"temperature,omitempty"`
	MaxTokens      int      `json:"max_tokens,omitempty"`
	TopP           float64  `json:"top_p,omitempty"`
	StopSequences  []string `json:"stop_sequences,omitempty"`
	JSONMode       bool     `json:"json_mode,omitempty"`
	EnableThinking bool     `json:"enable_thinking,omitempty"`
	ThinkingBudget int      `json:"thinking_budget,omitempty"`
}

// DefaultChatOptions returns sensible defaults for chat options.
func DefaultChatOptions() ChatOptions {
	return ChatOptions{
		Temperature: 0.7,
		MaxTokens:   4096,
		TopP:        1.0,
	}
}

// ChatRequest represents a unified request structure for all providers.
type ChatRequest struct {
	SystemPrompt string            `json:"system_prompt"`
	UserPrompt   string            `json:"user_prompt"`
	Messages     []Message         `json:"messages,omitempty"`
	Options      ChatOptions       `json:"options"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// ChatResponse represents a unified response structure from all providers.
type ChatResponse struct {
	// Response content separation (thinking vs content)
	Thinking string `json:"thinking,omitempty"`
	Content  string `json:"content"`

	// Accurate token measurement (extracted from provider API response)
	InputTokens    int `json:"input_tokens"`
	OutputTokens   int `json:"output_tokens"`
	ThinkingTokens int `json:"thinking_tokens,omitempty"`
	TotalTokens    int `json:"total_tokens"`

	// Metadata
	Provider  ProviderType `json:"provider"`
	Model     string       `json:"model"`
	LatencyMs int64        `json:"latency_ms"`
	RequestID string       `json:"request_id,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
}

// StreamDelta represents a streaming response chunk.
type StreamDelta struct {
	Type    string `json:"type"` // "thinking" or "content"
	Content string `json:"content"`
}

// StreamHandler provides callbacks for streaming responses.
type StreamHandler struct {
	OnThinking func(delta string)
	OnContent  func(delta string)
	OnError    func(err error)
	OnDone     func(response *ChatResponse)
}

// ProviderConfig holds configuration for a single provider.
type ProviderConfig struct {
	Type           ProviderType `json:"type"`
	APIKey         string       `json:"api_key"`
	Model          string       `json:"model"`
	Enabled        bool         `json:"enabled"`
	Priority       int          `json:"priority"`
	ThinkingBudget int          `json:"thinking_budget"`
}

// Config holds API keys for all providers.
// Provider/model selection is managed in DB (ai_configs table).
type Config struct {
	OpenAIAPIKey string
	GeminiAPIKey string
	ClaudeAPIKey string
}
