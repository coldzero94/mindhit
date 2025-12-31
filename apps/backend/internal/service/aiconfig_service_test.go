package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mindhit/api/internal/infrastructure/ai"
	"github.com/mindhit/api/internal/testutil"
)

func TestAIConfigService_GetConfigForTask(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	svc := NewAIConfigService(client)

	// Create a test config
	_, err := svc.Upsert(ctx, UpsertAIConfigRequest{
		TaskType:    "tag_extraction",
		Provider:    "gemini",
		Model:       "gemini-2.0-flash",
		Temperature: 0.3,
		MaxTokens:   1024,
		JSONMode:    true,
		Enabled:     true,
	})
	require.NoError(t, err)

	// Get config
	cfg, err := svc.GetConfigForTask(ctx, "tag_extraction")
	require.NoError(t, err)
	assert.Equal(t, "tag_extraction", cfg.TaskType)
	assert.Equal(t, "gemini", cfg.Provider)
	assert.Equal(t, "gemini-2.0-flash", cfg.Model)
	assert.True(t, cfg.JSONMode)
}

func TestAIConfigService_GetConfigForTask_FallbackToDefault(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	svc := NewAIConfigService(client)

	// Create default config only
	_, err := svc.Upsert(ctx, UpsertAIConfigRequest{
		TaskType:    "default",
		Provider:    "openai",
		Model:       "gpt-4o",
		Temperature: 0.7,
		MaxTokens:   4096,
		Enabled:     true,
	})
	require.NoError(t, err)

	// Get non-existent task type - should fall back to default
	cfg, err := svc.GetConfigForTask(ctx, "nonexistent_task")
	require.NoError(t, err)
	assert.Equal(t, "default", cfg.TaskType)
	assert.Equal(t, "openai", cfg.Provider)
}

func TestAIConfigService_GetConfigForTask_Caching(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	svc := NewAIConfigService(client)

	// Create config
	_, err := svc.Upsert(ctx, UpsertAIConfigRequest{
		TaskType:    "test_task",
		Provider:    "gemini",
		Model:       "gemini-2.0-flash",
		Temperature: 0.5,
		MaxTokens:   2048,
		Enabled:     true,
	})
	require.NoError(t, err)

	// First call - cache miss
	cfg1, err := svc.GetConfigForTask(ctx, "test_task")
	require.NoError(t, err)
	assert.Equal(t, "test_task", cfg1.TaskType)

	// Second call - should hit cache
	cfg2, err := svc.GetConfigForTask(ctx, "test_task")
	require.NoError(t, err)
	assert.Equal(t, cfg1.ID, cfg2.ID)
}

func TestAIConfigService_InvalidateCache(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	svc := NewAIConfigService(client)

	// Create and cache config
	_, err := svc.Upsert(ctx, UpsertAIConfigRequest{
		TaskType:    "cache_test",
		Provider:    "gemini",
		Model:       "gemini-2.0-flash",
		Temperature: 0.5,
		MaxTokens:   1024,
		Enabled:     true,
	})
	require.NoError(t, err)

	_, err = svc.GetConfigForTask(ctx, "cache_test")
	require.NoError(t, err)

	// Invalidate cache
	svc.InvalidateCache()

	// Verify cache is empty
	assert.Empty(t, svc.cache)
}

func TestAIConfigService_Upsert_Create(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	svc := NewAIConfigService(client)

	cfg, err := svc.Upsert(ctx, UpsertAIConfigRequest{
		TaskType:          "new_task",
		Provider:          "claude",
		Model:             "claude-sonnet-4-20250514",
		FallbackProviders: []string{"openai", "gemini"},
		Temperature:       0.5,
		MaxTokens:         8192,
		ThinkingBudget:    10000,
		JSONMode:          true,
		Enabled:           true,
		UpdatedBy:         "test",
	})

	require.NoError(t, err)
	assert.Equal(t, "new_task", cfg.TaskType)
	assert.Equal(t, "claude", cfg.Provider)
	assert.Equal(t, "claude-sonnet-4-20250514", cfg.Model)
	assert.Equal(t, []string{"openai", "gemini"}, cfg.FallbackProviders)
	assert.Equal(t, 10000, cfg.ThinkingBudget)
	assert.True(t, cfg.JSONMode)
}

func TestAIConfigService_Upsert_Update(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	svc := NewAIConfigService(client)

	// Create initial config
	_, err := svc.Upsert(ctx, UpsertAIConfigRequest{
		TaskType:    "update_task",
		Provider:    "openai",
		Model:       "gpt-4o",
		Temperature: 0.7,
		MaxTokens:   4096,
		Enabled:     true,
	})
	require.NoError(t, err)

	// Update config
	updated, err := svc.Upsert(ctx, UpsertAIConfigRequest{
		TaskType:    "update_task",
		Provider:    "gemini",
		Model:       "gemini-2.0-flash",
		Temperature: 0.3,
		MaxTokens:   2048,
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.Equal(t, "update_task", updated.TaskType)
	assert.Equal(t, "gemini", updated.Provider)
	assert.Equal(t, "gemini-2.0-flash", updated.Model)
}

func TestAIConfigService_Delete(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	svc := NewAIConfigService(client)

	// Create config
	_, err := svc.Upsert(ctx, UpsertAIConfigRequest{
		TaskType:    "delete_task",
		Provider:    "gemini",
		Model:       "gemini-2.0-flash",
		Temperature: 0.5,
		MaxTokens:   1024,
		Enabled:     true,
	})
	require.NoError(t, err)

	// Delete config
	err = svc.Delete(ctx, "delete_task")
	require.NoError(t, err)

	// Verify deleted
	cfgs, err := svc.GetAll(ctx)
	require.NoError(t, err)
	for _, cfg := range cfgs {
		assert.NotEqual(t, "delete_task", cfg.TaskType)
	}
}

func TestAIConfigService_GetAll(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	svc := NewAIConfigService(client)

	// Create multiple configs
	_, err := svc.Upsert(ctx, UpsertAIConfigRequest{
		TaskType: "task_a", Provider: "gemini", Model: "gemini-2.0-flash", Enabled: true,
	})
	require.NoError(t, err)

	_, err = svc.Upsert(ctx, UpsertAIConfigRequest{
		TaskType: "task_b", Provider: "openai", Model: "gpt-4o", Enabled: true,
	})
	require.NoError(t, err)

	// Get all
	cfgs, err := svc.GetAll(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(cfgs), 2)
}

func TestAIConfigService_SeedDefaultConfigs(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	svc := NewAIConfigService(client)

	// Seed defaults
	err := svc.SeedDefaultConfigs(ctx)
	require.NoError(t, err)

	// Verify defaults created
	cfgs, err := svc.GetAll(ctx)
	require.NoError(t, err)

	// Find specific configs
	taskTypes := make(map[string]bool)
	for _, cfg := range cfgs {
		taskTypes[cfg.TaskType] = true
	}
	assert.True(t, taskTypes["default"])
	assert.True(t, taskTypes[string(ai.TaskTagExtraction)])
	assert.True(t, taskTypes[string(ai.TaskMindmap)])

	// Seed again - should not create duplicates
	err = svc.SeedDefaultConfigs(ctx)
	require.NoError(t, err)

	cfgs2, err := svc.GetAll(ctx)
	require.NoError(t, err)
	// Count should remain the same after re-seeding
	assert.Equal(t, len(cfgs), len(cfgs2))
}
