package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/mindhit/api/ent"
	enturl "github.com/mindhit/api/ent/url"
	"github.com/mindhit/api/internal/infrastructure/ai"
	"github.com/mindhit/api/internal/infrastructure/metrics"
	"github.com/mindhit/api/internal/infrastructure/queue"
	"github.com/mindhit/api/internal/service"
)

const batchTagExtractionPrompt = `Analyze the following web pages and extract keywords and summaries for each.

%s

For each page, extract:
1. 3-5 core keywords (Korean nouns)
2. 1-2 sentence summary (Korean)

Respond in JSON format with an array:
[
  {
    "url_id": "<uuid from input>",
    "keywords": ["키워드1", "키워드2", "키워드3"],
    "summary": "페이지 요약"
  }
]`

// BatchTagResult represents the AI response for batch tag extraction.
type BatchTagResult struct {
	URLID    string   `json:"url_id"`
	Keywords []string `json:"keywords"`
	Summary  string   `json:"summary"`
}

// HandleURLBatchTagExtraction processes batch tag extraction for multiple URLs.
func (h *handlers) HandleURLBatchTagExtraction(ctx context.Context, t *asynq.Task) error {
	start := time.Now()
	jobType := "batch_tag_extraction"

	defer func() {
		metrics.WorkerJobDuration.WithLabelValues(jobType).Observe(time.Since(start).Seconds())
	}()

	var payload queue.URLBatchTagExtractionPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	if len(payload.URLIDs) == 0 {
		slog.Debug("empty url batch, skipping")
		return nil
	}

	slog.Info("batch extracting tags", "url_count", len(payload.URLIDs), "session_id", payload.SessionID)

	// Check if AI manager is available
	if h.aiManager == nil {
		slog.Warn("ai manager not configured, skipping batch tag extraction")
		return nil
	}

	// Parse user and session IDs for usage tracking
	var userID, sessionID uuid.UUID
	if payload.UserID != "" {
		userID, _ = uuid.Parse(payload.UserID)
	}
	if payload.SessionID != "" {
		sessionID, _ = uuid.Parse(payload.SessionID)
	}

	// Check usage limit before AI call
	if h.usageService != nil && userID != uuid.Nil {
		status, err := h.usageService.CheckLimit(ctx, userID)
		if err != nil {
			slog.Warn("failed to check usage limit", "error", err)
		} else if !status.CanUseAI {
			slog.Warn("user token limit exceeded, skipping batch tag extraction",
				"user_id", userID,
				"tokens_used", status.TokensUsed,
				"token_limit", status.TokenLimit,
			)
			return fmt.Errorf("token limit exceeded: used %d/%d", status.TokensUsed, status.TokenLimit)
		}
	}

	// Parse URL IDs
	urlIDs := make([]uuid.UUID, 0, len(payload.URLIDs))
	for _, id := range payload.URLIDs {
		uid, err := uuid.Parse(id)
		if err != nil {
			slog.Warn("invalid url id in batch", "url_id", id, "error", err)
			continue
		}
		urlIDs = append(urlIDs, uid)
	}

	if len(urlIDs) == 0 {
		return nil
	}

	// Fetch all URLs in single query
	urls, err := h.client.URL.Query().
		Where(enturl.IDIn(urlIDs...)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("fetch urls: %w", err)
	}

	// Filter URLs that need processing (no keywords and have content)
	urlsToProcess := make([]*ent.URL, 0)
	for _, u := range urls {
		if len(u.Keywords) > 0 {
			slog.Debug("url already has keywords, skipping", "url_id", u.ID)
			continue
		}
		if u.Content == "" {
			slog.Debug("url has no content, skipping", "url_id", u.ID)
			continue
		}
		urlsToProcess = append(urlsToProcess, u)
	}

	if len(urlsToProcess) == 0 {
		slog.Debug("no urls need processing in batch")
		return nil
	}

	// Build combined prompt
	var promptBuilder strings.Builder
	for i, u := range urlsToProcess {
		content := truncateContent(u.Content, 2000) // Smaller limit for batch
		promptBuilder.WriteString(fmt.Sprintf(`
### Page %d
- URL ID: %s
- Title: %s
- Content:
%s
`,
			i+1,
			u.ID.String(),
			u.Title,
			content,
		))
	}

	// Generate tags using AI
	req := ai.ChatRequest{
		UserPrompt: fmt.Sprintf(batchTagExtractionPrompt, promptBuilder.String()),
		Options: ai.ChatOptions{
			MaxTokens: 200 * len(urlsToProcess), // ~200 tokens per URL
			JSONMode:  true,
		},
		Metadata: map[string]string{
			"session_id": sessionID.String(),
			"user_id":    userID.String(),
			"batch_size": fmt.Sprintf("%d", len(urlsToProcess)),
		},
	}

	response, err := h.aiManager.Chat(ctx, ai.TaskTagExtraction, req)
	if err != nil {
		return fmt.Errorf("ai batch tag extraction: %w", err)
	}

	// Record token usage
	if h.usageService != nil && userID != uuid.Nil {
		if err := h.usageService.RecordUsage(ctx, service.UsageRecord{
			UserID:    userID,
			SessionID: sessionID,
			Operation: "batch_tag_extraction",
			Tokens:    response.TotalTokens,
			AIModel:   response.Model,
		}); err != nil {
			slog.Error("failed to record batch tag extraction usage", "error", err)
		}
	}

	// Parse response
	var results []BatchTagResult
	if err := json.Unmarshal([]byte(response.Content), &results); err != nil {
		return fmt.Errorf("parse ai response: %w", err)
	}

	// Update URLs with results
	successCount := 0
	for _, result := range results {
		urlID, err := uuid.Parse(result.URLID)
		if err != nil {
			slog.Warn("invalid url_id in response", "url_id", result.URLID)
			continue
		}

		_, err = h.client.URL.UpdateOneID(urlID).
			SetKeywords(result.Keywords).
			SetSummary(result.Summary).
			Save(ctx)
		if err != nil {
			slog.Error("failed to update url", "url_id", urlID, "error", err)
			continue
		}
		successCount++
	}

	metrics.WorkerJobsProcessed.WithLabelValues(jobType, "success").Inc()

	slog.Info("batch extracted tags",
		"batch_size", len(urlsToProcess),
		"success_count", successCount,
		"provider", response.Provider,
		"tokens", response.TotalTokens,
	)
	return nil
}
