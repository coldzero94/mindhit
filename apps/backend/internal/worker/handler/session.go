package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/mindhit/api/ent/pagevisit"
	"github.com/mindhit/api/ent/session"
	"github.com/mindhit/api/internal/infrastructure/queue"
)

// HandleSessionProcess processes a completed session.
func (h *handlers) HandleSessionProcess(ctx context.Context, t *asynq.Task) error {
	var payload queue.SessionProcessPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	slog.Info("processing session", "session_id", payload.SessionID)

	sessionID, err := uuid.Parse(payload.SessionID)
	if err != nil {
		return fmt.Errorf("invalid session ID: %w", err)
	}

	// Get session
	sess, err := h.client.Session.Query().
		Where(session.IDEQ(sessionID)).
		Only(ctx)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Verify session is in processing state
	if sess.SessionStatus != session.SessionStatusProcessing {
		slog.Warn("session not in processing state",
			"session_id", payload.SessionID,
			"status", sess.SessionStatus,
		)
		return nil // Not an error, just skip
	}

	// Get session with user for userID
	sess, err = h.client.Session.Query().
		Where(session.IDEQ(sessionID)).
		WithUser().
		Only(ctx)
	if err != nil {
		return fmt.Errorf("failed to get session with user: %w", err)
	}

	userID := sess.Edges.User.ID.String()

	// Get all URLs for this session via PageVisits
	pageVisits, err := h.client.PageVisit.Query().
		Where(pagevisit.HasSessionWith(session.IDEQ(sessionID))).
		WithURL().
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed to get page visits: %w", err)
	}

	// Collect unique URL IDs that need tag extraction (no keywords yet)
	urlIDSet := make(map[string]bool)
	var urlIDs []string
	for _, pv := range pageVisits {
		if pv.Edges.URL != nil {
			urlID := pv.Edges.URL.ID.String()
			// Only add URLs that don't have keywords yet
			if !urlIDSet[urlID] && len(pv.Edges.URL.Keywords) == 0 {
				urlIDSet[urlID] = true
				urlIDs = append(urlIDs, urlID)
			}
		}
	}

	slog.Info("found URLs for tag extraction",
		"session_id", payload.SessionID,
		"total_page_visits", len(pageVisits),
		"urls_to_process", len(urlIDs),
	)

	// Enqueue AI tasks if queue client is available
	if h.queueClient != nil {
		// Enqueue batch tag extraction tasks (5 URLs per batch)
		const batchSize = 5
		for i := 0; i < len(urlIDs); i += batchSize {
			end := i + batchSize
			if end > len(urlIDs) {
				end = len(urlIDs)
			}
			batch := urlIDs[i:end]

			task, err := queue.NewURLBatchTagExtractionTask(batch, payload.SessionID, userID)
			if err != nil {
				slog.Error("failed to create batch tag extraction task", "error", err)
				continue
			}

			_, err = h.queueClient.Enqueue(task, asynq.MaxRetry(3))
			if err != nil {
				slog.Error("failed to enqueue batch tag extraction task", "error", err)
				continue
			}

			slog.Info("enqueued batch tag extraction",
				"session_id", payload.SessionID,
				"batch_size", len(batch),
			)
		}

		// Enqueue mindmap generation task
		mindmapTask, err := queue.NewMindmapGenerateTask(payload.SessionID)
		if err != nil {
			slog.Error("failed to create mindmap task", "error", err)
		} else {
			// Delay mindmap generation to allow tag extraction to complete first
			_, err = h.queueClient.Enqueue(mindmapTask, asynq.MaxRetry(3), asynq.ProcessIn(30*time.Second))
			if err != nil {
				slog.Error("failed to enqueue mindmap task", "error", err)
			} else {
				slog.Info("enqueued mindmap generation", "session_id", payload.SessionID)
			}
		}
	} else {
		slog.Warn("queue client not configured, skipping AI task enqueueing")
	}

	// Mark as completed
	_, err = h.client.Session.UpdateOneID(sessionID).
		SetSessionStatus(session.SessionStatusCompleted).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to update session status: %w", err)
	}

	slog.Info("session processing completed", "session_id", payload.SessionID)
	return nil
}
