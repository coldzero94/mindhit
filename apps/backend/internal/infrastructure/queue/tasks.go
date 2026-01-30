package queue

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

// Task types
const (
	TypeSessionProcess        = "session:process"
	TypeSessionCleanup        = "session:cleanup"
	TypeURLSummarize          = "url:summarize"
	TypeURLTagExtraction      = "url:tag_extraction"
	TypeURLBatchTagExtraction = "url:batch_tag_extraction"
	TypeMindmapGenerate       = "mindmap:generate"
)

// SessionProcessPayload is the payload for session processing.
type SessionProcessPayload struct {
	SessionID string `json:"session_id"`
}

// NewSessionProcessTask creates a new session process task.
func NewSessionProcessTask(sessionID string) (*asynq.Task, error) {
	payload, err := json.Marshal(SessionProcessPayload{SessionID: sessionID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSessionProcess, payload), nil
}

// SessionCleanupPayload is the payload for session cleanup.
type SessionCleanupPayload struct {
	MaxAgeHours int `json:"max_age_hours"`
}

// NewSessionCleanupTask creates a new session cleanup task.
func NewSessionCleanupTask(maxAgeHours int) (*asynq.Task, error) {
	payload, err := json.Marshal(SessionCleanupPayload{MaxAgeHours: maxAgeHours})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSessionCleanup, payload), nil
}

// URLSummarizePayload is the payload for URL summarization.
type URLSummarizePayload struct {
	SessionID string `json:"session_id"`
	URL       string `json:"url"`
}

// NewURLSummarizeTask creates a new URL summarize task.
func NewURLSummarizeTask(sessionID, url string) (*asynq.Task, error) {
	payload, err := json.Marshal(URLSummarizePayload{
		SessionID: sessionID,
		URL:       url,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeURLSummarize, payload), nil
}

// URLTagExtractionPayload is the payload for URL tag extraction.
type URLTagExtractionPayload struct {
	URLID string `json:"url_id"`
}

// NewURLTagExtractionTask creates a new URL tag extraction task.
func NewURLTagExtractionTask(urlID string) (*asynq.Task, error) {
	payload, err := json.Marshal(URLTagExtractionPayload{URLID: urlID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeURLTagExtraction, payload), nil
}

// URLBatchTagExtractionPayload is the payload for batch URL tag extraction.
type URLBatchTagExtractionPayload struct {
	URLIDs    []string `json:"url_ids"`
	SessionID string   `json:"session_id"`
	UserID    string   `json:"user_id"`
}

// NewURLBatchTagExtractionTask creates a new batch URL tag extraction task.
func NewURLBatchTagExtractionTask(urlIDs []string, sessionID, userID string) (*asynq.Task, error) {
	payload, err := json.Marshal(URLBatchTagExtractionPayload{
		URLIDs:    urlIDs,
		SessionID: sessionID,
		UserID:    userID,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeURLBatchTagExtraction, payload), nil
}

// MindmapGeneratePayload is the payload for mindmap generation.
type MindmapGeneratePayload struct {
	SessionID string `json:"session_id"`
}

// NewMindmapGenerateTask creates a new mindmap generate task.
func NewMindmapGenerateTask(sessionID string) (*asynq.Task, error) {
	payload, err := json.Marshal(MindmapGeneratePayload{SessionID: sessionID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeMindmapGenerate, payload), nil
}
