package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/ent/session"
	"github.com/mindhit/api/internal/service"
	"github.com/mindhit/api/internal/testutil"
)

func setupSessionServiceTest(t *testing.T) (*ent.Client, *service.SessionService, *service.AuthService) {
	client := testutil.SetupTestDB(t)
	sessionService := service.NewSessionService(client)
	authService := service.NewAuthService(client)
	return client, sessionService, authService
}

// uniqueEmail generates a unique email for each test
func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s-%s@test.com", prefix, uuid.New().String()[:8])
}

func createTestUser(t *testing.T, authService *service.AuthService, email string) *ent.User {
	ctx := context.Background()
	user, err := authService.Signup(ctx, email, "password123")
	require.NoError(t, err)
	return user
}

// ==================== Start Tests ====================

func TestSessionService_Start_Success(t *testing.T) {
	client, sessionService, authService := setupSessionServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("start"))

	sess, err := sessionService.Start(ctx, user.ID)

	require.NoError(t, err)
	assert.NotNil(t, sess)
	assert.Equal(t, session.SessionStatusRecording, sess.SessionStatus)
	assert.False(t, sess.StartedAt.IsZero())
}

// ==================== Pause Tests ====================

func TestSessionService_Pause_Success(t *testing.T) {
	client, sessionService, authService := setupSessionServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("test"))

	sess, err := sessionService.Start(ctx, user.ID)
	require.NoError(t, err)

	pausedSess, err := sessionService.Pause(ctx, sess.ID, user.ID)

	require.NoError(t, err)
	assert.Equal(t, session.SessionStatusPaused, pausedSess.SessionStatus)
}

func TestSessionService_Pause_InvalidState(t *testing.T) {
	client, sessionService, authService := setupSessionServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("test"))

	sess, err := sessionService.Start(ctx, user.ID)
	require.NoError(t, err)

	// Pause once
	_, err = sessionService.Pause(ctx, sess.ID, user.ID)
	require.NoError(t, err)

	// Try to pause again (should fail)
	_, err = sessionService.Pause(ctx, sess.ID, user.ID)

	assert.ErrorIs(t, err, service.ErrInvalidSessionState)
}

func TestSessionService_Pause_NotOwned(t *testing.T) {
	client, sessionService, authService := setupSessionServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user1 := createTestUser(t, authService, uniqueEmail("user1"))
	user2 := createTestUser(t, authService, uniqueEmail("user2"))

	sess, err := sessionService.Start(ctx, user1.ID)
	require.NoError(t, err)

	_, err = sessionService.Pause(ctx, sess.ID, user2.ID)

	assert.ErrorIs(t, err, service.ErrSessionNotOwned)
}

func TestSessionService_Pause_NotFound(t *testing.T) {
	client, sessionService, authService := setupSessionServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("test"))

	_, err := sessionService.Pause(ctx, uuid.New(), user.ID)

	assert.ErrorIs(t, err, service.ErrSessionNotFound)
}

// ==================== Resume Tests ====================

func TestSessionService_Resume_Success(t *testing.T) {
	client, sessionService, authService := setupSessionServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("test"))

	sess, err := sessionService.Start(ctx, user.ID)
	require.NoError(t, err)

	_, err = sessionService.Pause(ctx, sess.ID, user.ID)
	require.NoError(t, err)

	resumedSess, err := sessionService.Resume(ctx, sess.ID, user.ID)

	require.NoError(t, err)
	assert.Equal(t, session.SessionStatusRecording, resumedSess.SessionStatus)
}

func TestSessionService_Resume_InvalidState(t *testing.T) {
	client, sessionService, authService := setupSessionServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("test"))

	sess, err := sessionService.Start(ctx, user.ID)
	require.NoError(t, err)

	// Try to resume a recording session (should fail)
	_, err = sessionService.Resume(ctx, sess.ID, user.ID)

	assert.ErrorIs(t, err, service.ErrInvalidSessionState)
}

// ==================== Stop Tests ====================

func TestSessionService_Stop_FromRecording(t *testing.T) {
	client, sessionService, authService := setupSessionServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("test"))

	sess, err := sessionService.Start(ctx, user.ID)
	require.NoError(t, err)

	stoppedSess, err := sessionService.Stop(ctx, sess.ID, user.ID)

	require.NoError(t, err)
	assert.Equal(t, session.SessionStatusProcessing, stoppedSess.SessionStatus)
	assert.NotNil(t, stoppedSess.EndedAt)
}

func TestSessionService_Stop_FromPaused(t *testing.T) {
	client, sessionService, authService := setupSessionServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("test"))

	sess, err := sessionService.Start(ctx, user.ID)
	require.NoError(t, err)

	_, err = sessionService.Pause(ctx, sess.ID, user.ID)
	require.NoError(t, err)

	stoppedSess, err := sessionService.Stop(ctx, sess.ID, user.ID)

	require.NoError(t, err)
	assert.Equal(t, session.SessionStatusProcessing, stoppedSess.SessionStatus)
}

func TestSessionService_Stop_InvalidState(t *testing.T) {
	client, sessionService, authService := setupSessionServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("test"))

	sess, err := sessionService.Start(ctx, user.ID)
	require.NoError(t, err)

	// Stop the session
	_, err = sessionService.Stop(ctx, sess.ID, user.ID)
	require.NoError(t, err)

	// Try to stop again (should fail)
	_, err = sessionService.Stop(ctx, sess.ID, user.ID)

	assert.ErrorIs(t, err, service.ErrInvalidSessionState)
}

// ==================== Get Tests ====================

func TestSessionService_Get_Success(t *testing.T) {
	client, sessionService, authService := setupSessionServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("test"))

	sess, err := sessionService.Start(ctx, user.ID)
	require.NoError(t, err)

	retrieved, err := sessionService.Get(ctx, sess.ID, user.ID)

	require.NoError(t, err)
	assert.Equal(t, sess.ID, retrieved.ID)
}

func TestSessionService_Get_NotFound(t *testing.T) {
	client, sessionService, authService := setupSessionServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("test"))

	_, err := sessionService.Get(ctx, uuid.New(), user.ID)

	assert.ErrorIs(t, err, service.ErrSessionNotFound)
}

func TestSessionService_Get_NotOwned(t *testing.T) {
	client, sessionService, authService := setupSessionServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user1 := createTestUser(t, authService, uniqueEmail("user1"))
	user2 := createTestUser(t, authService, uniqueEmail("user2"))

	sess, err := sessionService.Start(ctx, user1.ID)
	require.NoError(t, err)

	_, err = sessionService.Get(ctx, sess.ID, user2.ID)

	assert.ErrorIs(t, err, service.ErrSessionNotOwned)
}

// ==================== ListByUser Tests ====================

func TestSessionService_ListByUser_Success(t *testing.T) {
	client, sessionService, authService := setupSessionServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("test"))

	// Create multiple sessions
	_, err := sessionService.Start(ctx, user.ID)
	require.NoError(t, err)
	_, err = sessionService.Start(ctx, user.ID)
	require.NoError(t, err)
	_, err = sessionService.Start(ctx, user.ID)
	require.NoError(t, err)

	sessions, err := sessionService.ListByUser(ctx, user.ID, 10, 0)

	require.NoError(t, err)
	assert.Len(t, sessions, 3)
}

func TestSessionService_ListByUser_Pagination(t *testing.T) {
	client, sessionService, authService := setupSessionServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("test"))

	// Create 5 sessions
	for i := 0; i < 5; i++ {
		_, err := sessionService.Start(ctx, user.ID)
		require.NoError(t, err)
	}

	// Get first 2
	sessions, err := sessionService.ListByUser(ctx, user.ID, 2, 0)
	require.NoError(t, err)
	assert.Len(t, sessions, 2)

	// Get next 2
	sessions, err = sessionService.ListByUser(ctx, user.ID, 2, 2)
	require.NoError(t, err)
	assert.Len(t, sessions, 2)

	// Get last 1
	sessions, err = sessionService.ListByUser(ctx, user.ID, 2, 4)
	require.NoError(t, err)
	assert.Len(t, sessions, 1)
}

func TestSessionService_ListByUser_Empty(t *testing.T) {
	client, sessionService, authService := setupSessionServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("test"))

	sessions, err := sessionService.ListByUser(ctx, user.ID, 10, 0)

	require.NoError(t, err)
	assert.Empty(t, sessions)
}

func TestSessionService_ListByUser_IsolatedByUser(t *testing.T) {
	client, sessionService, authService := setupSessionServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user1 := createTestUser(t, authService, uniqueEmail("user1"))
	user2 := createTestUser(t, authService, uniqueEmail("user2"))

	// Create sessions for user1
	_, err := sessionService.Start(ctx, user1.ID)
	require.NoError(t, err)
	_, err = sessionService.Start(ctx, user1.ID)
	require.NoError(t, err)

	// Create session for user2
	_, err = sessionService.Start(ctx, user2.ID)
	require.NoError(t, err)

	// User1 should only see their sessions
	sessions, err := sessionService.ListByUser(ctx, user1.ID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, sessions, 2)

	// User2 should only see their session
	sessions, err = sessionService.ListByUser(ctx, user2.ID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, sessions, 1)
}

// ==================== Update Tests ====================

func TestSessionService_Update_Title(t *testing.T) {
	client, sessionService, authService := setupSessionServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("test"))

	sess, err := sessionService.Start(ctx, user.ID)
	require.NoError(t, err)

	title := "My Research Session"
	updatedSess, err := sessionService.Update(ctx, sess.ID, user.ID, &title, nil)

	require.NoError(t, err)
	assert.Equal(t, title, *updatedSess.Title)
}

func TestSessionService_Update_Description(t *testing.T) {
	client, sessionService, authService := setupSessionServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("test"))

	sess, err := sessionService.Start(ctx, user.ID)
	require.NoError(t, err)

	description := "Researching AI topics"
	updatedSess, err := sessionService.Update(ctx, sess.ID, user.ID, nil, &description)

	require.NoError(t, err)
	assert.Equal(t, description, *updatedSess.Description)
}

func TestSessionService_Update_Both(t *testing.T) {
	client, sessionService, authService := setupSessionServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("test"))

	sess, err := sessionService.Start(ctx, user.ID)
	require.NoError(t, err)

	title := "My Research Session"
	description := "Researching AI topics"
	updatedSess, err := sessionService.Update(ctx, sess.ID, user.ID, &title, &description)

	require.NoError(t, err)
	assert.Equal(t, title, *updatedSess.Title)
	assert.Equal(t, description, *updatedSess.Description)
}

func TestSessionService_Update_NotOwned(t *testing.T) {
	client, sessionService, authService := setupSessionServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user1 := createTestUser(t, authService, uniqueEmail("user1"))
	user2 := createTestUser(t, authService, uniqueEmail("user2"))

	sess, err := sessionService.Start(ctx, user1.ID)
	require.NoError(t, err)

	title := "Hacked Title"
	_, err = sessionService.Update(ctx, sess.ID, user2.ID, &title, nil)

	assert.ErrorIs(t, err, service.ErrSessionNotOwned)
}

// ==================== Delete Tests ====================

func TestSessionService_Delete_Success(t *testing.T) {
	client, sessionService, authService := setupSessionServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("test"))

	sess, err := sessionService.Start(ctx, user.ID)
	require.NoError(t, err)

	err = sessionService.Delete(ctx, sess.ID, user.ID)

	require.NoError(t, err)

	// Session should not be found after deletion
	_, err = sessionService.Get(ctx, sess.ID, user.ID)
	assert.ErrorIs(t, err, service.ErrSessionNotFound)
}

func TestSessionService_Delete_NotFound(t *testing.T) {
	client, sessionService, authService := setupSessionServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user := createTestUser(t, authService, uniqueEmail("test"))

	err := sessionService.Delete(ctx, uuid.New(), user.ID)

	assert.ErrorIs(t, err, service.ErrSessionNotFound)
}

func TestSessionService_Delete_NotOwned(t *testing.T) {
	client, sessionService, authService := setupSessionServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	user1 := createTestUser(t, authService, uniqueEmail("user1"))
	user2 := createTestUser(t, authService, uniqueEmail("user2"))

	sess, err := sessionService.Start(ctx, user1.ID)
	require.NoError(t, err)

	err = sessionService.Delete(ctx, sess.ID, user2.ID)

	assert.ErrorIs(t, err, service.ErrSessionNotOwned)
}

// ==================== State Transition Tests ====================

func TestSessionService_StateTransitions(t *testing.T) {
	tests := []struct {
		name          string
		setupState    func(t *testing.T, svc *service.SessionService, sessionID, userID uuid.UUID)
		action        func(svc *service.SessionService, sessionID, userID uuid.UUID) error
		expectedError error
	}{
		{
			name: "recording -> paused (valid)",
			setupState: func(t *testing.T, _ *service.SessionService, _ uuid.UUID, _ uuid.UUID) {
				t.Helper() // Already in recording state - no setup needed
			},
			action: func(svc *service.SessionService, sessionID, userID uuid.UUID) error {
				_, err := svc.Pause(context.Background(), sessionID, userID)
				return err
			},
			expectedError: nil,
		},
		{
			name: "recording -> resume (invalid)",
			setupState: func(t *testing.T, _ *service.SessionService, _ uuid.UUID, _ uuid.UUID) {
				t.Helper() // Already in recording state - no setup needed
			},
			action: func(svc *service.SessionService, sessionID, userID uuid.UUID) error {
				_, err := svc.Resume(context.Background(), sessionID, userID)
				return err
			},
			expectedError: service.ErrInvalidSessionState,
		},
		{
			name: "paused -> recording (valid)",
			setupState: func(t *testing.T, svc *service.SessionService, sessionID, userID uuid.UUID) {
				_, err := svc.Pause(context.Background(), sessionID, userID)
				require.NoError(t, err)
			},
			action: func(svc *service.SessionService, sessionID, userID uuid.UUID) error {
				_, err := svc.Resume(context.Background(), sessionID, userID)
				return err
			},
			expectedError: nil,
		},
		{
			name: "paused -> paused (invalid)",
			setupState: func(t *testing.T, svc *service.SessionService, sessionID, userID uuid.UUID) {
				_, err := svc.Pause(context.Background(), sessionID, userID)
				require.NoError(t, err)
			},
			action: func(svc *service.SessionService, sessionID, userID uuid.UUID) error {
				_, err := svc.Pause(context.Background(), sessionID, userID)
				return err
			},
			expectedError: service.ErrInvalidSessionState,
		},
		{
			name: "processing -> pause (invalid)",
			setupState: func(t *testing.T, svc *service.SessionService, sessionID, userID uuid.UUID) {
				_, err := svc.Stop(context.Background(), sessionID, userID)
				require.NoError(t, err)
			},
			action: func(svc *service.SessionService, sessionID, userID uuid.UUID) error {
				_, err := svc.Pause(context.Background(), sessionID, userID)
				return err
			},
			expectedError: service.ErrInvalidSessionState,
		},
		{
			name: "processing -> resume (invalid)",
			setupState: func(t *testing.T, svc *service.SessionService, sessionID, userID uuid.UUID) {
				_, err := svc.Stop(context.Background(), sessionID, userID)
				require.NoError(t, err)
			},
			action: func(svc *service.SessionService, sessionID, userID uuid.UUID) error {
				_, err := svc.Resume(context.Background(), sessionID, userID)
				return err
			},
			expectedError: service.ErrInvalidSessionState,
		},
		{
			name: "processing -> stop (invalid)",
			setupState: func(t *testing.T, svc *service.SessionService, sessionID, userID uuid.UUID) {
				_, err := svc.Stop(context.Background(), sessionID, userID)
				require.NoError(t, err)
			},
			action: func(svc *service.SessionService, sessionID, userID uuid.UUID) error {
				_, err := svc.Stop(context.Background(), sessionID, userID)
				return err
			},
			expectedError: service.ErrInvalidSessionState,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, sessionService, authService := setupSessionServiceTest(t)
			defer testutil.CleanupTestDB(t, client)

			ctx := context.Background()
			user := createTestUser(t, authService, uniqueEmail("test"))

			sess, err := sessionService.Start(ctx, user.ID)
			require.NoError(t, err)

			tt.setupState(t, sessionService, sess.ID, user.ID)
			err = tt.action(sessionService, sess.ID, user.ID)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
