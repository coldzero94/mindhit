package service

import (
	"context"

	"github.com/mindhit/api/internal/infrastructure/ai"
)

// AILogAdapter adapts AILogService to implement ai.LogProvider interface.
type AILogAdapter struct {
	service *AILogService
}

// NewAILogAdapter creates a new AILogAdapter.
func NewAILogAdapter(service *AILogService) *AILogAdapter {
	return &AILogAdapter{service: service}
}

// Log implements ai.LogProvider interface.
func (a *AILogAdapter) Log(ctx context.Context, req ai.LogRequest) error {
	_, err := a.service.Log(ctx, AILogRequest{
		UserID:       req.UserID,
		SessionID:    req.SessionID,
		TaskType:     req.TaskType,
		Request:      req.Request,
		Response:     req.Response,
		ErrorMessage: req.ErrorMessage,
	})
	return err
}
