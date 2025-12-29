package ai

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/sashabaranov/go-openai"
)

// OpenAIProvider implements the Provider interface for OpenAI.
type OpenAIProvider struct {
	BaseProvider
	client *openai.Client
}

// NewOpenAIProvider creates a new OpenAI provider.
func NewOpenAIProvider(cfg ProviderConfig) *OpenAIProvider {
	model := cfg.Model
	if model == "" {
		model = "gpt-4o"
	}
	return &OpenAIProvider{
		BaseProvider: BaseProvider{
			providerType:   ProviderOpenAI,
			model:          model,
			thinkingBudget: cfg.ThinkingBudget,
		},
		client: openai.NewClient(cfg.APIKey),
	}
}

// Chat sends a chat completion request to OpenAI.
func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	startTime := time.Now()
	messages := buildMessages(req)

	chatMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		chatMessages[i] = openai.ChatCompletionMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
	}

	apiReq := openai.ChatCompletionRequest{
		Model:       p.model,
		Messages:    chatMessages,
		Temperature: float32(req.Options.Temperature),
		MaxTokens:   req.Options.MaxTokens,
		TopP:        float32(req.Options.TopP),
		Stop:        req.Options.StopSequences,
	}

	// JSON mode
	if req.Options.JSONMode {
		apiReq.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		}
	}

	resp, err := p.client.CreateChatCompletion(ctx, apiReq)
	if err != nil {
		return nil, fmt.Errorf("openai chat: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, ErrNoResponse
	}

	content := resp.Choices[0].Message.Content

	// Validate JSON if JSON mode enabled
	if err := validateJSONResponse(content, req.Options.JSONMode); err != nil {
		return nil, err
	}

	return &ChatResponse{
		Content:      content,
		Thinking:     "",
		Provider:     ProviderOpenAI,
		Model:        resp.Model,
		InputTokens:  resp.Usage.PromptTokens,
		OutputTokens: resp.Usage.CompletionTokens,
		TotalTokens:  resp.Usage.TotalTokens,
		LatencyMs:    time.Since(startTime).Milliseconds(),
		RequestID:    resp.ID,
		CreatedAt:    time.Now(),
	}, nil
}

// ChatStream sends a streaming chat completion request to OpenAI.
func (p *OpenAIProvider) ChatStream(ctx context.Context, req ChatRequest, handler StreamHandler) error {
	startTime := time.Now()
	messages := buildMessages(req)

	chatMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		chatMessages[i] = openai.ChatCompletionMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
	}

	stream, err := p.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:       p.model,
		Messages:    chatMessages,
		Temperature: float32(req.Options.Temperature),
		MaxTokens:   req.Options.MaxTokens,
		Stream:      true,
	})
	if err != nil {
		if handler.OnError != nil {
			handler.OnError(err)
		}
		return err
	}
	defer func() { _ = stream.Close() }()

	var fullContent string
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			if handler.OnError != nil {
				handler.OnError(err)
			}
			return err
		}
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta.Content
			fullContent += delta
			if handler.OnContent != nil {
				handler.OnContent(delta)
			}
		}
	}

	if handler.OnDone != nil {
		handler.OnDone(&ChatResponse{
			Content:   fullContent,
			Provider:  ProviderOpenAI,
			Model:     p.model,
			LatencyMs: time.Since(startTime).Milliseconds(),
			CreatedAt: time.Now(),
		})
	}
	return nil
}

// IsHealthy checks if the OpenAI provider is available.
func (p *OpenAIProvider) IsHealthy(ctx context.Context) bool {
	_, err := p.Chat(ctx, ChatRequest{
		UserPrompt: "ping",
		Options:    ChatOptions{MaxTokens: 5},
	})
	if err != nil {
		slog.Warn("openai health check failed", "error", err)
		return false
	}
	return true
}
