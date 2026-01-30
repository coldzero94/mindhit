package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/ent/pagevisit"
	"github.com/mindhit/api/ent/url"
	"github.com/mindhit/api/internal/infrastructure/ai"
	"github.com/mindhit/api/internal/infrastructure/metrics"
	"github.com/mindhit/api/internal/infrastructure/queue"
	"github.com/mindhit/api/internal/service"
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

	// Record token usage - find user via URL -> PageVisit -> Session -> User
	if h.usageService != nil {
		h.recordTagExtractionUsage(ctx, urlID, response)
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

// recordTagExtractionUsage finds the user who visited this URL and records token usage.
func (h *handlers) recordTagExtractionUsage(ctx context.Context, urlID uuid.UUID, response *ai.ChatResponse) {
	// Find a PageVisit that references this URL to get session/user info
	pv, err := h.client.PageVisit.
		Query().
		Where(pagevisit.HasURLWith(url.IDEQ(urlID))).
		WithSession(func(q *ent.SessionQuery) {
			q.WithUser()
		}).
		First(ctx)

	if err != nil {
		slog.Debug("could not find pagevisit for url, skipping usage recording", "url_id", urlID, "error", err)
		return
	}

	if pv.Edges.Session == nil || pv.Edges.Session.Edges.User == nil {
		slog.Debug("pagevisit has no session or user, skipping usage recording", "url_id", urlID)
		return
	}

	userID := pv.Edges.Session.Edges.User.ID
	sessionID := pv.Edges.Session.ID

	if err := h.usageService.RecordUsage(ctx, service.UsageRecord{
		UserID:    userID,
		SessionID: sessionID,
		Operation: "tag_extraction",
		Tokens:    response.TotalTokens,
		AIModel:   response.Model,
	}); err != nil {
		slog.Error("failed to record tag extraction usage", "error", err)
	}
}
