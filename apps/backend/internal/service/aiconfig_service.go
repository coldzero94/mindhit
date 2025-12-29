package service

import (
	"context"
	"sync"
	"time"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/ent/aiconfig"
)

// AIConfigService manages AI provider configuration in DB with caching.
type AIConfigService struct {
	client *ent.Client

	// In-memory cache
	cache     map[string]*ent.AIConfig
	cacheMu   sync.RWMutex
	cacheTime time.Time
	cacheTTL  time.Duration
}

// NewAIConfigService creates a new AIConfigService.
func NewAIConfigService(client *ent.Client) *AIConfigService {
	return &AIConfigService{
		client:   client,
		cache:    make(map[string]*ent.AIConfig),
		cacheTTL: 5 * time.Minute,
	}
}

// GetConfigForTask returns the config for a specific task type.
func (s *AIConfigService) GetConfigForTask(ctx context.Context, taskType string) (*ent.AIConfig, error) {
	s.cacheMu.RLock()
	if time.Since(s.cacheTime) < s.cacheTTL {
		if cfg, ok := s.cache[taskType]; ok {
			s.cacheMu.RUnlock()
			return cfg, nil
		}
	}
	s.cacheMu.RUnlock()

	return s.refreshCache(ctx, taskType)
}

// refreshCache loads config from DB and updates cache.
func (s *AIConfigService) refreshCache(ctx context.Context, taskType string) (*ent.AIConfig, error) {
	cfg, err := s.client.AIConfig.Query().
		Where(aiconfig.TaskTypeEQ(taskType)).
		Where(aiconfig.EnabledEQ(true)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) && taskType != "default" {
			return s.refreshCache(ctx, "default")
		}
		return nil, err
	}

	s.cacheMu.Lock()
	s.cache[taskType] = cfg
	s.cacheTime = time.Now()
	s.cacheMu.Unlock()

	return cfg, nil
}

// InvalidateCache clears the cache.
func (s *AIConfigService) InvalidateCache() {
	s.cacheMu.Lock()
	s.cache = make(map[string]*ent.AIConfig)
	s.cacheTime = time.Time{}
	s.cacheMu.Unlock()
}

// GetAll returns all AI configs.
func (s *AIConfigService) GetAll(ctx context.Context) ([]*ent.AIConfig, error) {
	return s.client.AIConfig.Query().
		Order(ent.Asc(aiconfig.FieldTaskType)).
		All(ctx)
}

// UpsertAIConfigRequest is the request for creating/updating AI config.
type UpsertAIConfigRequest struct {
	TaskType          string   `json:"task_type"`
	Provider          string   `json:"provider"`
	Model             string   `json:"model"`
	FallbackProviders []string `json:"fallback_providers,omitempty"`
	Temperature       float64  `json:"temperature"`
	MaxTokens         int      `json:"max_tokens"`
	ThinkingBudget    int      `json:"thinking_budget,omitempty"`
	JSONMode          bool     `json:"json_mode"`
	Enabled           bool     `json:"enabled"`
	UpdatedBy         string   `json:"updated_by,omitempty"`
}

// Upsert creates or updates an AI config.
func (s *AIConfigService) Upsert(ctx context.Context, req UpsertAIConfigRequest) (*ent.AIConfig, error) {
	// Check if config exists
	existing, err := s.client.AIConfig.Query().
		Where(aiconfig.TaskTypeEQ(req.TaskType)).
		Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}

	var cfg *ent.AIConfig
	if existing != nil {
		// Update existing
		cfg, err = existing.Update().
			SetProvider(req.Provider).
			SetModel(req.Model).
			SetFallbackProviders(req.FallbackProviders).
			SetTemperature(req.Temperature).
			SetMaxTokens(req.MaxTokens).
			SetThinkingBudget(req.ThinkingBudget).
			SetJSONMode(req.JSONMode).
			SetEnabled(req.Enabled).
			SetUpdatedBy(req.UpdatedBy).
			Save(ctx)
	} else {
		// Create new
		cfg, err = s.client.AIConfig.Create().
			SetTaskType(req.TaskType).
			SetProvider(req.Provider).
			SetModel(req.Model).
			SetFallbackProviders(req.FallbackProviders).
			SetTemperature(req.Temperature).
			SetMaxTokens(req.MaxTokens).
			SetThinkingBudget(req.ThinkingBudget).
			SetJSONMode(req.JSONMode).
			SetEnabled(req.Enabled).
			SetUpdatedBy(req.UpdatedBy).
			Save(ctx)
	}

	if err != nil {
		return nil, err
	}

	s.InvalidateCache()
	return cfg, nil
}

// Delete removes an AI config.
func (s *AIConfigService) Delete(ctx context.Context, taskType string) error {
	_, err := s.client.AIConfig.Delete().
		Where(aiconfig.TaskTypeEQ(taskType)).
		Exec(ctx)

	if err == nil {
		s.InvalidateCache()
	}
	return err
}

// SeedDefaultConfigs creates default AI configs if they don't exist.
func (s *AIConfigService) SeedDefaultConfigs(ctx context.Context) error {
	defaults := []UpsertAIConfigRequest{
		{
			TaskType:          "default",
			Provider:          "openai",
			Model:             "gpt-4o",
			FallbackProviders: []string{"gemini", "claude"},
			Temperature:       0.7,
			MaxTokens:         4096,
			Enabled:           true,
		},
		{
			TaskType:          "tag_extraction",
			Provider:          "gemini",
			Model:             "gemini-2.0-flash",
			FallbackProviders: []string{"openai"},
			Temperature:       0.3,
			MaxTokens:         1024,
			JSONMode:          true,
			Enabled:           true,
		},
		{
			TaskType:          "mindmap",
			Provider:          "claude",
			Model:             "claude-sonnet-4-20250514",
			FallbackProviders: []string{"openai"},
			Temperature:       0.5,
			MaxTokens:         8192,
			ThinkingBudget:    10000,
			JSONMode:          true,
			Enabled:           true,
		},
	}

	for _, cfg := range defaults {
		exists, _ := s.client.AIConfig.Query().
			Where(aiconfig.TaskTypeEQ(cfg.TaskType)).
			Exist(ctx)
		if !exists {
			if _, err := s.Upsert(ctx, cfg); err != nil {
				return err
			}
		}
	}

	return nil
}
