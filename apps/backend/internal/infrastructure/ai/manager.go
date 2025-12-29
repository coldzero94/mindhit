package ai

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"

	"github.com/mindhit/api/ent"
)

// ConfigProvider provides AI configuration from database.
type ConfigProvider interface {
	GetConfigForTask(ctx context.Context, taskType string) (*ent.AIConfig, error)
}

// LogProvider logs AI requests to the database.
type LogProvider interface {
	Log(ctx context.Context, req LogRequest) error
}

// LogRequest represents the data needed to create an AI log entry.
type LogRequest struct {
	UserID       *uuid.UUID
	SessionID    *uuid.UUID
	TaskType     TaskType
	Request      ChatRequest
	Response     *ChatResponse
	ErrorMessage string
}

// ProviderManager manages AI providers with DB-based config and auto-logging.
type ProviderManager struct {
	providers      map[ProviderType]Provider
	configProvider ConfigProvider
	logProvider    LogProvider
	mu             sync.RWMutex
}

// NewProviderManager creates a ProviderManager with environment API keys.
func NewProviderManager(
	ctx context.Context,
	cfg Config,
	configProvider ConfigProvider,
	logProvider LogProvider,
) (*ProviderManager, error) {
	pm := &ProviderManager{
		providers:      make(map[ProviderType]Provider),
		configProvider: configProvider,
		logProvider:    logProvider,
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
		slog.Warn("no ai providers configured (missing API keys)")
	} else {
		slog.Info("provider manager initialized", "available_providers", len(pm.providers))
	}

	return pm, nil
}

// Chat executes a request using DB-configured provider for the task.
func (pm *ProviderManager) Chat(ctx context.Context, task TaskType, req ChatRequest) (*ChatResponse, error) {
	cfg, err := pm.configProvider.GetConfigForTask(ctx, string(task))
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
func (pm *ProviderManager) getProvidersFromConfig(cfg *ent.AIConfig) []Provider {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var result []Provider
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
	if pm.logProvider == nil {
		return
	}

	logReq := LogRequest{
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

	if err := pm.logProvider.Log(ctx, logReq); err != nil {
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

// HasProviders returns true if at least one provider is configured.
func (pm *ProviderManager) HasProviders() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.providers) > 0
}
