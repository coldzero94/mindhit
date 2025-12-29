package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// ClaudeProvider implements the Provider interface for Anthropic Claude.
type ClaudeProvider struct {
	BaseProvider
	client anthropic.Client
}

// NewClaudeProvider creates a new Anthropic Claude provider.
func NewClaudeProvider(cfg ProviderConfig) *ClaudeProvider {
	model := cfg.Model
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	client := anthropic.NewClient(
		option.WithAPIKey(cfg.APIKey),
	)

	return &ClaudeProvider{
		BaseProvider: BaseProvider{
			providerType:   ProviderClaude,
			model:          model,
			thinkingBudget: cfg.ThinkingBudget,
		},
		client: client,
	}
}

// Chat sends a chat completion request to Anthropic Claude.
func (p *ClaudeProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	startTime := time.Now()
	messages := buildMessages(req)

	var anthropicMessages []anthropic.MessageParam
	for _, msg := range messages {
		switch msg.Role {
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
		Model:     anthropic.Model(p.model),
		MaxTokens: int64(req.Options.MaxTokens),
		Messages:  anthropicMessages,
	}

	// System prompt
	if req.SystemPrompt != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: req.SystemPrompt},
		}
	}

	// Temperature
	if req.Options.Temperature > 0 {
		params.Temperature = anthropic.Float(req.Options.Temperature)
	}

	// TopP
	if req.Options.TopP > 0 && req.Options.TopP < 1 {
		params.TopP = anthropic.Float(req.Options.TopP)
	}

	resp, err := p.client.Messages.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("claude message: %w", err)
	}

	if len(resp.Content) == 0 {
		return nil, ErrNoResponse
	}

	// Extract content from response
	var thinking, content string
	for _, block := range resp.Content {
		switch block.Type {
		case "thinking":
			thinking = block.Thinking
		case "text":
			content += block.Text
		}
	}

	// Validate JSON if requested
	if req.Options.JSONMode {
		var js json.RawMessage
		if err := json.Unmarshal([]byte(content), &js); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
		}
	}

	return &ChatResponse{
		Content:      content,
		Thinking:     thinking,
		Provider:     ProviderClaude,
		Model:        string(resp.Model),
		InputTokens:  int(resp.Usage.InputTokens),
		OutputTokens: int(resp.Usage.OutputTokens),
		TotalTokens:  int(resp.Usage.InputTokens + resp.Usage.OutputTokens),
		LatencyMs:    time.Since(startTime).Milliseconds(),
		RequestID:    resp.ID,
		CreatedAt:    time.Now(),
	}, nil
}

// ChatStream sends a streaming chat completion request to Anthropic Claude.
func (p *ClaudeProvider) ChatStream(ctx context.Context, req ChatRequest, handler StreamHandler) error {
	startTime := time.Now()
	messages := buildMessages(req)

	var anthropicMessages []anthropic.MessageParam
	for _, msg := range messages {
		switch msg.Role {
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
		Model:     anthropic.Model(p.model),
		MaxTokens: int64(req.Options.MaxTokens),
		Messages:  anthropicMessages,
	}

	if req.SystemPrompt != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: req.SystemPrompt},
		}
	}

	stream := p.client.Messages.NewStreaming(ctx, params)

	var fullThinking, fullContent string
	for stream.Next() {
		event := stream.Current()
		if event.Type == "content_block_delta" {
			delta := event.Delta
			if delta.Type == "thinking_delta" {
				fullThinking += delta.Thinking
				if handler.OnThinking != nil {
					handler.OnThinking(delta.Thinking)
				}
			} else if delta.Type == "text_delta" {
				fullContent += delta.Text
				if handler.OnContent != nil {
					handler.OnContent(delta.Text)
				}
			}
		}
	}

	if err := stream.Err(); err != nil {
		if handler.OnError != nil {
			handler.OnError(err)
		}
		return err
	}

	if handler.OnDone != nil {
		handler.OnDone(&ChatResponse{
			Content:   fullContent,
			Thinking:  fullThinking,
			Provider:  ProviderClaude,
			Model:     p.model,
			LatencyMs: time.Since(startTime).Milliseconds(),
			CreatedAt: time.Now(),
		})
	}
	return nil
}

// IsHealthy checks if the Claude provider is available.
func (p *ClaudeProvider) IsHealthy(ctx context.Context) bool {
	_, err := p.Chat(ctx, ChatRequest{
		UserPrompt: "ping",
		Options:    ChatOptions{MaxTokens: 5},
	})
	return err == nil
}
