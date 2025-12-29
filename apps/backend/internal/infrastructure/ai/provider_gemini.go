package ai

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// GeminiProvider implements the Provider interface for Google Gemini.
type GeminiProvider struct {
	BaseProvider
	client *genai.Client
}

// NewGeminiProvider creates a new Google Gemini provider.
func NewGeminiProvider(ctx context.Context, cfg ProviderConfig) (*GeminiProvider, error) {
	model := cfg.Model
	if model == "" {
		model = "gemini-2.0-flash"
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.APIKey))
	if err != nil {
		return nil, fmt.Errorf("create gemini client: %w", err)
	}

	return &GeminiProvider{
		BaseProvider: BaseProvider{
			providerType:   ProviderGemini,
			model:          model,
			thinkingBudget: cfg.ThinkingBudget,
		},
		client: client,
	}, nil
}

// Chat sends a chat completion request to Google Gemini.
func (p *GeminiProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	startTime := time.Now()
	model := p.client.GenerativeModel(p.model)

	model.SetTemperature(float32(req.Options.Temperature))
	model.SetMaxOutputTokens(int32(req.Options.MaxTokens))
	model.SetTopP(float32(req.Options.TopP))

	if len(req.Options.StopSequences) > 0 {
		model.StopSequences = req.Options.StopSequences
	}

	// JSON mode
	if req.Options.JSONMode {
		model.ResponseMIMEType = "application/json"
	}

	// System prompt
	if req.SystemPrompt != "" {
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(req.SystemPrompt)},
		}
	}

	// Build parts from messages + user prompt
	var parts []genai.Part
	for _, msg := range req.Messages {
		if msg.Role != RoleSystem {
			parts = append(parts, genai.Text(msg.Content))
		}
	}
	if req.UserPrompt != "" {
		parts = append(parts, genai.Text(req.UserPrompt))
	}

	resp, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return nil, fmt.Errorf("gemini generate: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, ErrNoResponse
	}

	var content strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			content.WriteString(string(text))
		}
	}

	// Validate JSON if requested
	if err := validateJSONResponse(content.String(), req.Options.JSONMode); err != nil {
		return nil, err
	}

	var inputTokens, outputTokens, totalTokens int
	if resp.UsageMetadata != nil {
		inputTokens = int(resp.UsageMetadata.PromptTokenCount)
		outputTokens = int(resp.UsageMetadata.CandidatesTokenCount)
		totalTokens = int(resp.UsageMetadata.TotalTokenCount)
	}

	return &ChatResponse{
		Content:      content.String(),
		Provider:     ProviderGemini,
		Model:        p.model,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalTokens:  totalTokens,
		LatencyMs:    time.Since(startTime).Milliseconds(),
		CreatedAt:    time.Now(),
	}, nil
}

// ChatStream sends a streaming chat completion request to Google Gemini.
func (p *GeminiProvider) ChatStream(ctx context.Context, req ChatRequest, handler StreamHandler) error {
	startTime := time.Now()
	model := p.client.GenerativeModel(p.model)

	model.SetTemperature(float32(req.Options.Temperature))
	model.SetMaxOutputTokens(int32(req.Options.MaxTokens))

	if req.SystemPrompt != "" {
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(req.SystemPrompt)},
		}
	}

	var parts []genai.Part
	if req.UserPrompt != "" {
		parts = append(parts, genai.Text(req.UserPrompt))
	}

	iter := model.GenerateContentStream(ctx, parts...)
	var fullContent strings.Builder

	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			if handler.OnError != nil {
				handler.OnError(err)
			}
			return err
		}
		for _, cand := range resp.Candidates {
			if cand.Content == nil {
				continue
			}
			for _, part := range cand.Content.Parts {
				if text, ok := part.(genai.Text); ok {
					delta := string(text)
					fullContent.WriteString(delta)
					if handler.OnContent != nil {
						handler.OnContent(delta)
					}
				}
			}
		}
	}

	if handler.OnDone != nil {
		handler.OnDone(&ChatResponse{
			Content:   fullContent.String(),
			Provider:  ProviderGemini,
			Model:     p.model,
			LatencyMs: time.Since(startTime).Milliseconds(),
			CreatedAt: time.Now(),
		})
	}
	return nil
}

// IsHealthy checks if the Gemini provider is available.
func (p *GeminiProvider) IsHealthy(ctx context.Context) bool {
	_, err := p.Chat(ctx, ChatRequest{
		UserPrompt: "ping",
		Options:    ChatOptions{MaxTokens: 5},
	})
	if err != nil {
		slog.Warn("gemini health check failed", "error", err)
		return false
	}
	return true
}

// Close closes the Gemini client.
func (p *GeminiProvider) Close() error {
	return p.client.Close()
}
