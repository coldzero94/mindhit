package service

import (
	"context"
	"sync"
	"time"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/ent/aiconfig"
	"github.com/mindhit/api/internal/infrastructure/ai"
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
	// Try read lock first for cache hit
	s.cacheMu.RLock()
	if time.Since(s.cacheTime) < s.cacheTTL {
		if cfg, ok := s.cache[taskType]; ok {
			s.cacheMu.RUnlock()
			return cfg, nil
		}
	}
	s.cacheMu.RUnlock()

	// Acquire write lock for cache refresh
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	// Double-check after acquiring write lock (another goroutine may have refreshed)
	if time.Since(s.cacheTime) < s.cacheTTL {
		if cfg, ok := s.cache[taskType]; ok {
			return cfg, nil
		}
	}

	return s.refreshCacheLocked(ctx, taskType)
}

// refreshCacheLocked loads config from DB and updates cache.
// Caller must hold s.cacheMu write lock.
func (s *AIConfigService) refreshCacheLocked(ctx context.Context, taskType string) (*ent.AIConfig, error) {
	cfg, err := s.client.AIConfig.Query().
		Where(aiconfig.TaskTypeEQ(taskType)).
		Where(aiconfig.EnabledEQ(true)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) && taskType != "default" {
			return s.refreshCacheLocked(ctx, "default")
		}
		return nil, err
	}

	s.cache[taskType] = cfg
	s.cacheTime = time.Now()

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
			Provider:          string(ai.ProviderOpenAI),
			Model:             ai.DefaultOpenAIModel,
			FallbackProviders: []string{string(ai.ProviderGemini), string(ai.ProviderClaude)},
			Temperature:       0.7,
			MaxTokens:         4096,
			Enabled:           true,
		},
		{
			TaskType:          string(ai.TaskTagExtraction),
			Provider:          string(ai.ProviderGemini),
			Model:             ai.DefaultGeminiModel,
			FallbackProviders: []string{string(ai.ProviderOpenAI)},
			Temperature:       0.3,
			MaxTokens:         1024,
			JSONMode:          true,
			Enabled:           true,
		},
		{
			TaskType:          string(ai.TaskMindmap),
			Provider:          string(ai.ProviderClaude),
			Model:             ai.DefaultClaudeModel,
			FallbackProviders: []string{string(ai.ProviderOpenAI)},
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
