package service

import (
	"context"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/internal/infrastructure/ai"
)

// AIConfigAdapter adapts AIConfigService to implement ai.ConfigProvider interface.
type AIConfigAdapter struct {
	service *AIConfigService
}

// NewAIConfigAdapter creates a new AIConfigAdapter.
func NewAIConfigAdapter(service *AIConfigService) *AIConfigAdapter {
	return &AIConfigAdapter{service: service}
}

// GetConfigForTask implements ai.ConfigProvider interface.
func (a *AIConfigAdapter) GetConfigForTask(ctx context.Context, taskType string) (*ent.AIConfig, error) {
	return a.service.GetConfigForTask(ctx, taskType)
}

// Compile-time check that adapters implement the interfaces.
var (
	_ ai.ConfigProvider = (*AIConfigAdapter)(nil)
	_ ai.LogProvider    = (*AILogAdapter)(nil)
)
