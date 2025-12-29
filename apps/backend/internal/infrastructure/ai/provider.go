package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// Common errors for AI providers.
var (
	ErrProviderNotConfigured = errors.New("ai provider not configured")
	ErrNoResponse            = errors.New("no response from ai provider")
	ErrInvalidJSON           = errors.New("invalid json response")
)

// validateJSONResponse validates that the content is valid JSON when JSONMode is enabled.
func validateJSONResponse(content string, jsonMode bool) error {
	if !jsonMode {
		return nil
	}
	var js json.RawMessage
	if err := json.Unmarshal([]byte(content), &js); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}
	return nil
}

// Provider defines the interface that all AI providers must implement.
type Provider interface {
	// Chat sends a request and returns a unified response.
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)

	// ChatStream sends a request and streams the response.
	ChatStream(ctx context.Context, req ChatRequest, handler StreamHandler) error

	// Type returns the provider type.
	Type() ProviderType

	// Model returns the current model being used.
	Model() string

	// IsHealthy checks if the provider is available.
	IsHealthy(ctx context.Context) bool

	// Close releases any resources.
	Close() error
}

// BaseProvider contains common functionality for all providers.
type BaseProvider struct {
	providerType   ProviderType
	model          string
	thinkingBudget int
}

// Type returns the provider type.
func (b *BaseProvider) Type() ProviderType {
	return b.providerType
}

// Model returns the current model.
func (b *BaseProvider) Model() string {
	return b.model
}

// Close releases resources (default no-op).
func (b *BaseProvider) Close() error {
	return nil
}

// buildMessages converts ChatRequest to []Message for providers.
func buildMessages(req ChatRequest) []Message {
	var messages []Message

	// Add system prompt if present
	if req.SystemPrompt != "" {
		messages = append(messages, Message{
			Role:    RoleSystem,
			Content: req.SystemPrompt,
		})
	}

	// Add existing messages
	messages = append(messages, req.Messages...)

	// Add user prompt if present
	if req.UserPrompt != "" {
		messages = append(messages, Message{
			Role:    RoleUser,
			Content: req.UserPrompt,
		})
	}

	return messages
}
