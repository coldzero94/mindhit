package handler

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mindhit/api/ent/session"
	"github.com/mindhit/api/internal/infrastructure/queue"
	"github.com/mindhit/api/internal/testutil"
)

func TestHandleSessionProcess_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	h := &handlers{client: client}

	// Create test user with unique email
	user, err := client.User.Create().
		SetEmail("test-process-" + uuid.New().String() + "@example.com").
		SetPasswordHash("hashed").
		Save(ctx)
	require.NoError(t, err)

	// Create test session in processing state
	sess, err := client.Session.Create().
		SetUserID(user.ID).
		SetSessionStatus(session.SessionStatusProcessing).
		SetStartedAt(time.Now()).
		Save(ctx)
	require.NoError(t, err)

	// Create task payload
	payload, _ := json.Marshal(queue.SessionProcessPayload{SessionID: sess.ID.String()})
	task := asynq.NewTask(queue.TypeSessionProcess, payload)

	// Handle task
	err = h.HandleSessionProcess(ctx, task)

	require.NoError(t, err)

	// Verify session is now completed
	updated, err := client.Session.Get(ctx, sess.ID)
	require.NoError(t, err)
	assert.Equal(t, session.SessionStatusCompleted, updated.SessionStatus)
}

func TestHandleSessionProcess_NotInProcessingState(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	h := &handlers{client: client}

	// Create test user with unique email
	user, err := client.User.Create().
		SetEmail("test-skip-" + uuid.New().String() + "@example.com").
		SetPasswordHash("hashed").
		Save(ctx)
	require.NoError(t, err)

	// Create test session in recording state (not processing)
	sess, err := client.Session.Create().
		SetUserID(user.ID).
		SetSessionStatus(session.SessionStatusRecording).
		SetStartedAt(time.Now()).
		Save(ctx)
	require.NoError(t, err)

	// Create task payload
	payload, _ := json.Marshal(queue.SessionProcessPayload{SessionID: sess.ID.String()})
	task := asynq.NewTask(queue.TypeSessionProcess, payload)

	// Handle task - should skip without error
	err = h.HandleSessionProcess(ctx, task)

	require.NoError(t, err)

	// Verify session status unchanged
	updated, err := client.Session.Get(ctx, sess.ID)
	require.NoError(t, err)
	assert.Equal(t, session.SessionStatusRecording, updated.SessionStatus)
}

func TestHandleSessionProcess_InvalidPayload(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	h := &handlers{client: client}

	// Create task with invalid payload
	task := asynq.NewTask(queue.TypeSessionProcess, []byte("invalid json"))

	err := h.HandleSessionProcess(ctx, task)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal payload")
}

func TestHandleSessionProcess_SessionNotFound(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	h := &handlers{client: client}

	// Create task with non-existent session ID
	payload, _ := json.Marshal(queue.SessionProcessPayload{SessionID: uuid.New().String()})
	task := asynq.NewTask(queue.TypeSessionProcess, payload)

	err := h.HandleSessionProcess(ctx, task)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get session")
}

func TestHandleSessionProcess_InvalidSessionID(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	h := &handlers{client: client}

	// Create task with invalid UUID
	payload, _ := json.Marshal(queue.SessionProcessPayload{SessionID: "invalid-uuid"})
	task := asynq.NewTask(queue.TypeSessionProcess, payload)

	err := h.HandleSessionProcess(ctx, task)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid session ID")
}
