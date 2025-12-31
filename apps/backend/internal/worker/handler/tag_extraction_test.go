package handler

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"

	"github.com/mindhit/api/internal/infrastructure/queue"
	"github.com/mindhit/api/internal/testutil"
)

func TestHandleURLTagExtraction_NoAIManager(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	h := &handlers{
		client:    client,
		aiManager: nil, // No AI manager
	}

	// Create task payload
	payload, _ := json.Marshal(queue.URLTagExtractionPayload{URLID: uuid.New().String()})
	task := asynq.NewTask(queue.TypeURLTagExtraction, payload)

	// Should return nil (skip) when AI manager is not configured
	err := h.HandleURLTagExtraction(ctx, task)
	assert.NoError(t, err)
}

func TestHandleURLTagExtraction_InvalidPayload(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	h := &handlers{client: client}

	// Create task with invalid payload
	task := asynq.NewTask(queue.TypeURLTagExtraction, []byte("invalid json"))

	err := h.HandleURLTagExtraction(ctx, task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal payload")
}

func TestHandleURLTagExtraction_InvalidUUID(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	h := &handlers{client: client}

	// Create task with invalid UUID
	payload, _ := json.Marshal(queue.URLTagExtractionPayload{URLID: "not-a-uuid"})
	task := asynq.NewTask(queue.TypeURLTagExtraction, payload)

	err := h.HandleURLTagExtraction(ctx, task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse url id")
}

func TestHandleURLTagExtraction_URLNotFound(t *testing.T) {
	// We need a mock AI manager, but since we don't have URL, it will fail first
	// For now, skip this test as it requires a mock
	t.Skip("Requires AI manager mock")
}

func TestTruncateContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		maxLen   int
		expected string
	}{
		{
			name:     "content shorter than max",
			content:  "short",
			maxLen:   100,
			expected: "short",
		},
		{
			name:     "content exactly max length",
			content:  "12345",
			maxLen:   5,
			expected: "12345",
		},
		{
			name:     "content longer than max",
			content:  "1234567890",
			maxLen:   5,
			expected: "12345...",
		},
		{
			name:     "empty content",
			content:  "",
			maxLen:   10,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateContent(tt.content, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}
