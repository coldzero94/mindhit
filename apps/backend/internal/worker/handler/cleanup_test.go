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

func TestHandleSessionCleanup_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	h := &handlers{client: client}

	// Create test user with unique email
	user, err := client.User.Create().
		SetEmail("test-cleanup-" + uuid.New().String() + "@example.com").
		SetPasswordHash("hashed").
		Save(ctx)
	require.NoError(t, err)

	// Create old stale session (recording but updated 2 days ago)
	oldTime := time.Now().Add(-48 * time.Hour)
	staleSess, err := client.Session.Create().
		SetUserID(user.ID).
		SetSessionStatus(session.SessionStatusRecording).
		SetStartedAt(oldTime).
		Save(ctx)
	require.NoError(t, err)

	// Manually update the updated_at to make it stale
	_, err = client.Session.UpdateOneID(staleSess.ID).
		SetUpdatedAt(oldTime).
		Save(ctx)
	require.NoError(t, err)

	// Create recent session (should not be cleaned up)
	recentSess, err := client.Session.Create().
		SetUserID(user.ID).
		SetSessionStatus(session.SessionStatusRecording).
		SetStartedAt(time.Now()).
		Save(ctx)
	require.NoError(t, err)

	// Create task payload (24 hour max age)
	payload, _ := json.Marshal(queue.SessionCleanupPayload{MaxAgeHours: 24})
	task := asynq.NewTask(queue.TypeSessionCleanup, payload)

	// Handle task
	err = h.HandleSessionCleanup(ctx, task)

	require.NoError(t, err)

	// Verify stale session is now failed
	updatedStale, err := client.Session.Get(ctx, staleSess.ID)
	require.NoError(t, err)
	assert.Equal(t, session.SessionStatusFailed, updatedStale.SessionStatus)

	// Verify recent session is unchanged
	updatedRecent, err := client.Session.Get(ctx, recentSess.ID)
	require.NoError(t, err)
	assert.Equal(t, session.SessionStatusRecording, updatedRecent.SessionStatus)
}

func TestHandleSessionCleanup_InvalidPayload(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	h := &handlers{client: client}

	// Create task with invalid payload
	task := asynq.NewTask(queue.TypeSessionCleanup, []byte("invalid json"))

	err := h.HandleSessionCleanup(ctx, task)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal payload")
}

func TestHandleSessionCleanup_PausedSessions(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	h := &handlers{client: client}

	// Create test user
	user, err := client.User.Create().
		SetEmail("test-cleanup-paused-" + uuid.New().String() + "@example.com").
		SetPasswordHash("hashed").
		Save(ctx)
	require.NoError(t, err)

	// Create old paused session
	oldTime := time.Now().Add(-48 * time.Hour)
	pausedSess, err := client.Session.Create().
		SetUserID(user.ID).
		SetSessionStatus(session.SessionStatusPaused).
		SetStartedAt(oldTime).
		Save(ctx)
	require.NoError(t, err)

	// Manually update the updated_at to make it stale
	_, err = client.Session.UpdateOneID(pausedSess.ID).
		SetUpdatedAt(oldTime).
		Save(ctx)
	require.NoError(t, err)

	// Create task payload
	payload, _ := json.Marshal(queue.SessionCleanupPayload{MaxAgeHours: 24})
	task := asynq.NewTask(queue.TypeSessionCleanup, payload)

	// Handle task
	err = h.HandleSessionCleanup(ctx, task)

	require.NoError(t, err)

	// Verify paused session is now failed
	updated, err := client.Session.Get(ctx, pausedSess.ID)
	require.NoError(t, err)
	assert.Equal(t, session.SessionStatusFailed, updated.SessionStatus)
}

func TestHandleSessionCleanup_CompletedSessionsNotAffected(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	h := &handlers{client: client}

	// Create test user
	user, err := client.User.Create().
		SetEmail("test-cleanup-completed-" + uuid.New().String() + "@example.com").
		SetPasswordHash("hashed").
		Save(ctx)
	require.NoError(t, err)

	// Create old completed session (should NOT be cleaned up)
	oldTime := time.Now().Add(-48 * time.Hour)
	completedSess, err := client.Session.Create().
		SetUserID(user.ID).
		SetSessionStatus(session.SessionStatusCompleted).
		SetStartedAt(oldTime).
		Save(ctx)
	require.NoError(t, err)

	// Manually update the updated_at
	_, err = client.Session.UpdateOneID(completedSess.ID).
		SetUpdatedAt(oldTime).
		Save(ctx)
	require.NoError(t, err)

	// Create task payload
	payload, _ := json.Marshal(queue.SessionCleanupPayload{MaxAgeHours: 24})
	task := asynq.NewTask(queue.TypeSessionCleanup, payload)

	// Handle task
	err = h.HandleSessionCleanup(ctx, task)

	require.NoError(t, err)

	// Verify completed session is still completed
	updated, err := client.Session.Get(ctx, completedSess.ID)
	require.NoError(t, err)
	assert.Equal(t, session.SessionStatusCompleted, updated.SessionStatus)
}

func TestHandleSessionCleanup_NoStaleSessions(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	h := &handlers{client: client}

	// Create test user
	user, err := client.User.Create().
		SetEmail("test-cleanup-none-" + uuid.New().String() + "@example.com").
		SetPasswordHash("hashed").
		Save(ctx)
	require.NoError(t, err)

	// Create recent session
	recentSess, err := client.Session.Create().
		SetUserID(user.ID).
		SetSessionStatus(session.SessionStatusRecording).
		SetStartedAt(time.Now()).
		Save(ctx)
	require.NoError(t, err)

	// Create task payload
	payload, _ := json.Marshal(queue.SessionCleanupPayload{MaxAgeHours: 24})
	task := asynq.NewTask(queue.TypeSessionCleanup, payload)

	// Handle task
	err = h.HandleSessionCleanup(ctx, task)

	require.NoError(t, err)

	// Verify session is unchanged
	updated, err := client.Session.Get(ctx, recentSess.ID)
	require.NoError(t, err)
	assert.Equal(t, session.SessionStatusRecording, updated.SessionStatus)
}
