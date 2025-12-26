package queue

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSessionProcessTask(t *testing.T) {
	sessionID := "test-session-123"

	task, err := NewSessionProcessTask(sessionID)

	require.NoError(t, err)
	assert.Equal(t, TypeSessionProcess, task.Type())

	var payload SessionProcessPayload
	err = json.Unmarshal(task.Payload(), &payload)
	require.NoError(t, err)
	assert.Equal(t, sessionID, payload.SessionID)
}

func TestNewSessionCleanupTask(t *testing.T) {
	maxAgeHours := 24

	task, err := NewSessionCleanupTask(maxAgeHours)

	require.NoError(t, err)
	assert.Equal(t, TypeSessionCleanup, task.Type())

	var payload SessionCleanupPayload
	err = json.Unmarshal(task.Payload(), &payload)
	require.NoError(t, err)
	assert.Equal(t, maxAgeHours, payload.MaxAgeHours)
}

func TestNewURLSummarizeTask(t *testing.T) {
	sessionID := "session-456"
	url := "https://example.com/page"

	task, err := NewURLSummarizeTask(sessionID, url)

	require.NoError(t, err)
	assert.Equal(t, TypeURLSummarize, task.Type())

	var payload URLSummarizePayload
	err = json.Unmarshal(task.Payload(), &payload)
	require.NoError(t, err)
	assert.Equal(t, sessionID, payload.SessionID)
	assert.Equal(t, url, payload.URL)
}

func TestNewMindmapGenerateTask(t *testing.T) {
	sessionID := "session-789"

	task, err := NewMindmapGenerateTask(sessionID)

	require.NoError(t, err)
	assert.Equal(t, TypeMindmapGenerate, task.Type())

	var payload MindmapGeneratePayload
	err = json.Unmarshal(task.Payload(), &payload)
	require.NoError(t, err)
	assert.Equal(t, sessionID, payload.SessionID)
}

func TestTaskTypes(t *testing.T) {
	assert.Equal(t, "session:process", TypeSessionProcess)
	assert.Equal(t, "session:cleanup", TypeSessionCleanup)
	assert.Equal(t, "url:summarize", TypeURLSummarize)
	assert.Equal(t, "mindmap:generate", TypeMindmapGenerate)
}
