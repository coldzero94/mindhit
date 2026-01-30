package ai

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

const groqBaseURL = "https://api.groq.com/openai/v1"

// GroqProvider implements the Provider interface for Groq.
type GroqProvider struct {
	BaseProvider
	client *openai.Client
}

// NewGroqProvider creates a new Groq provider.
func NewGroqProvider(cfg ProviderConfig) *GroqProvider {
	model := cfg.Model
	if model == "" {
		model = DefaultGroqModel
	}

	// Create custom config with Groq's base URL
	clientConfig := openai.DefaultConfig(cfg.APIKey)
	clientConfig.BaseURL = groqBaseURL

	return &GroqProvider{
		BaseProvider: BaseProvider{
			providerType:   ProviderGroq,
			model:          model,
			thinkingBudget: cfg.ThinkingBudget,
		},
		client: openai.NewClientWithConfig(clientConfig),
	}
}

// Chat sends a chat completion request to Groq.
func (p *GroqProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
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
		return nil, fmt.Errorf("groq chat: %w", err)
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
		Provider:     ProviderGroq,
		Model:        resp.Model,
		InputTokens:  resp.Usage.PromptTokens,
		OutputTokens: resp.Usage.CompletionTokens,
		TotalTokens:  resp.Usage.TotalTokens,
		LatencyMs:    time.Since(startTime).Milliseconds(),
		RequestID:    resp.ID,
		CreatedAt:    time.Now(),
	}, nil
}

// ChatStream sends a streaming chat completion request to Groq.
func (p *GroqProvider) ChatStream(ctx context.Context, req ChatRequest, handler StreamHandler) error {
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

	var fullContent strings.Builder
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
			fullContent.WriteString(delta)
			if handler.OnContent != nil {
				handler.OnContent(delta)
			}
		}
	}

	if handler.OnDone != nil {
		handler.OnDone(&ChatResponse{
			Content:   fullContent.String(),
			Provider:  ProviderGroq,
			Model:     p.model,
			LatencyMs: time.Since(startTime).Milliseconds(),
			CreatedAt: time.Now(),
		})
	}
	return nil
}

// IsHealthy checks if the Groq provider is available.
func (p *GroqProvider) IsHealthy(ctx context.Context) bool {
	_, err := p.Chat(ctx, ChatRequest{
		UserPrompt: "ping",
		Options:    ChatOptions{MaxTokens: 5},
	})
	if err != nil {
		slog.Warn("groq health check failed", "error", err)
		return false
	}
	return true
}
