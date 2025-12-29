// Package service contains business logic services.
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
	status := ailog.StatusSuccess
	if req.ErrorMessage != "" {
		status = ailog.StatusError
	}

	// Handle nil response case (error logging)
	var (
		provider     = ""
		model        = ""
		content      = ""
		inputTokens  = 0
		outputTokens = 0
		thinkingTkns = 0
		totalTokens  = 0
		latencyMs    int64
		requestID    = ""
		thinking     = ""
	)

	if req.Response != nil {
		provider = string(req.Response.Provider)
		model = req.Response.Model
		content = req.Response.Content
		inputTokens = req.Response.InputTokens
		outputTokens = req.Response.OutputTokens
		thinkingTkns = req.Response.ThinkingTokens
		totalTokens = req.Response.TotalTokens
		latencyMs = req.Response.LatencyMs
		requestID = req.Response.RequestID
		thinking = req.Response.Thinking
	}

	builder := s.client.AILog.Create().
		SetTaskType(string(req.TaskType)).
		SetProvider(provider).
		SetModel(model).
		SetContent(content).
		SetInputTokens(inputTokens).
		SetOutputTokens(outputTokens).
		SetThinkingTokens(thinkingTkns).
		SetTotalTokens(totalTokens).
		SetLatencyMs(latencyMs).
		SetStatus(status)

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
	if thinking != "" {
		builder.SetThinking(thinking)
	}
	if requestID != "" {
		builder.SetRequestID(requestID)
	}
	if req.ErrorMessage != "" {
		builder.SetErrorMessage(req.ErrorMessage)
	}

	// Calculate estimated cost
	if req.Response != nil {
		cost := s.estimateCost(req.Response)
		builder.SetEstimatedCostCents(cost)
	}

	return builder.Save(ctx)
}

// estimateCost calculates cost in cents based on provider and tokens.
func (s *AILogService) estimateCost(resp *ai.ChatResponse) int {
	// Pricing per 1M tokens (approximate, in cents)
	pricing := map[ai.ProviderType]struct{ input, output int }{
		ai.ProviderOpenAI: {input: 250, output: 1000}, // GPT-4o
		ai.ProviderClaude: {input: 300, output: 1500}, // Claude 3.5 Sonnet
		ai.ProviderGemini: {input: 35, output: 105},   // Gemini 1.5 Flash
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
	logs, err := s.client.AILog.Query().
		Where(ailog.UserIDEQ(userID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	stats := &UsageStats{RequestCount: len(logs)}
	for _, log := range logs {
		stats.TotalTokens += log.TotalTokens
		stats.TotalCost += log.EstimatedCostCents
	}

	return stats, nil
}
