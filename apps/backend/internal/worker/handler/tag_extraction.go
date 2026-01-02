package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/mindhit/api/internal/infrastructure/ai"
	"github.com/mindhit/api/internal/infrastructure/metrics"
	"github.com/mindhit/api/internal/infrastructure/queue"
)

const tagExtractionPrompt = `Analyze the web page and extract the following:

1. Core keywords 3-5 (Korean nouns)
2. 1-2 sentence summary (Korean)

Page title: %s
Page content:
%s

Respond in JSON format:
{
  "keywords": ["키워드1", "키워드2", "키워드3"],
  "summary": "페이지 요약"
}`

// TagResult represents the AI response for tag extraction.
type TagResult struct {
	Keywords []string `json:"keywords"`
	Summary  string   `json:"summary"`
}

// HandleURLTagExtraction processes tag extraction for a URL.
func (h *handlers) HandleURLTagExtraction(ctx context.Context, t *asynq.Task) error {
	start := time.Now()
	jobType := "tag_extraction"

	defer func() {
		metrics.WorkerJobDuration.WithLabelValues(jobType).Observe(time.Since(start).Seconds())
	}()

	var payload queue.URLTagExtractionPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	urlID, err := uuid.Parse(payload.URLID)
	if err != nil {
		return fmt.Errorf("parse url id: %w", err)
	}

	slog.Info("extracting tags", "url_id", payload.URLID)

	// Check if AI manager is available
	if h.aiManager == nil {
		slog.Warn("ai manager not configured, skipping tag extraction")
		return nil
	}

	// Get URL from database
	u, err := h.client.URL.Get(ctx, urlID)
	if err != nil {
		return fmt.Errorf("get url: %w", err)
	}

	// Skip if already has keywords
	if len(u.Keywords) > 0 {
		slog.Debug("url already has keywords, skipping", "url", u.URL)
		return nil
	}

	// Skip if no content
	if u.Content == "" {
		slog.Warn("url has no content, skipping", "url", u.URL)
		return nil
	}

	// Generate tags using AI
	req := ai.ChatRequest{
		UserPrompt: fmt.Sprintf(tagExtractionPrompt, u.Title, truncateContent(u.Content, 8000)),
		Options: ai.ChatOptions{
			MaxTokens: 500,
			JSONMode:  true,
		},
		Metadata: map[string]string{
			"url_id": urlID.String(),
		},
	}

	response, err := h.aiManager.Chat(ctx, ai.TaskTagExtraction, req)
	if err != nil {
		return fmt.Errorf("ai tag extraction: %w", err)
	}

	var result TagResult
	if err := json.Unmarshal([]byte(response.Content), &result); err != nil {
		return fmt.Errorf("parse ai response: %w", err)
	}

	// Update URL with keywords and summary
	_, err = h.client.URL.UpdateOneID(urlID).
		SetKeywords(result.Keywords).
		SetSummary(result.Summary).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("update url: %w", err)
	}

	metrics.WorkerJobsProcessed.WithLabelValues(jobType, "success").Inc()

	slog.Info("extracted tags",
		"url", u.URL,
		"keywords", result.Keywords,
		"provider", response.Provider,
		"tokens", response.TotalTokens,
	)
	return nil
}

func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "..."
}
