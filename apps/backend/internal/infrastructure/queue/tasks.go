package queue

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

// Task types
const (
	TypeSessionProcess  = "session:process"
	TypeSessionCleanup  = "session:cleanup"
	TypeURLSummarize    = "url:summarize"
	TypeMindmapGenerate = "mindmap:generate"
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
