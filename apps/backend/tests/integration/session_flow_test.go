//go:build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mindhit/api/internal/controller"
	"github.com/mindhit/api/internal/generated"
	"github.com/mindhit/api/internal/service"
	"github.com/mindhit/api/internal/testutil"
)

const testAPIKey = "test-api-secret-key"

func getAuthHeader() string {
	// Use env var if set (for CI), otherwise use default
	key := os.Getenv("API_SECRET_KEY")
	if key == "" {
		key = testAPIKey
	}
	return "Bearer " + key
}

// TestSessionFlow_CreatePauseResumeStop tests the complete session lifecycle:
// 1. Start a session
// 2. Pause the session
// 3. Resume the session
// 4. Stop the session
// 5. Session appears in list
func TestSessionFlow_CreatePauseResumeStop(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	// Setup services
	sessionService := service.NewSessionService(client, nil) // nil queue for test
	sessionController := controller.NewSessionController(sessionService)

	ctx := context.Background()
	authHeader := getAuthHeader()

	// Step 1: Start session
	var sessionID string
	t.Run("Step 1: Start session", func(t *testing.T) {
		req := generated.RoutesStartRequestObject{
			Params: generated.RoutesStartParams{
				Authorization: authHeader,
			},
		}

		resp, err := sessionController.RoutesStart(ctx, req)
		require.NoError(t, err)

		startResp, ok := resp.(generated.RoutesStart201JSONResponse)
		require.True(t, ok, "expected 201 response")
		assert.Equal(t, generated.SessionSessionStatusRecording, startResp.Session.SessionStatus)

		sessionID = startResp.Session.Id
	})

	// Step 2: Pause session
	t.Run("Step 2: Pause session", func(t *testing.T) {
		req := generated.RoutesPauseRequestObject{
			Params: generated.RoutesPauseParams{
				Authorization: authHeader,
			},
			Id: sessionID,
		}

		resp, err := sessionController.RoutesPause(ctx, req)
		require.NoError(t, err)

		pauseResp, ok := resp.(generated.RoutesPause200JSONResponse)
		require.True(t, ok, "expected 200 response")
		assert.Equal(t, generated.SessionSessionStatusPaused, pauseResp.Session.SessionStatus)
	})

	// Step 3: Resume session
	t.Run("Step 3: Resume session", func(t *testing.T) {
		req := generated.RoutesResumeRequestObject{
			Params: generated.RoutesResumeParams{
				Authorization: authHeader,
			},
			Id: sessionID,
		}

		resp, err := sessionController.RoutesResume(ctx, req)
		require.NoError(t, err)

		resumeResp, ok := resp.(generated.RoutesResume200JSONResponse)
		require.True(t, ok, "expected 200 response")
		assert.Equal(t, generated.SessionSessionStatusRecording, resumeResp.Session.SessionStatus)
	})

	// Step 4: Stop session
	t.Run("Step 4: Stop session", func(t *testing.T) {
		req := generated.RoutesStopRequestObject{
			Params: generated.RoutesStopParams{
				Authorization: authHeader,
			},
			Id: sessionID,
		}

		resp, err := sessionController.RoutesStop(ctx, req)
		require.NoError(t, err)

		// Stop returns 200 with session in processing/completed state
		stopResp, ok := resp.(generated.RoutesStop200JSONResponse)
		require.True(t, ok, "expected 200 response")
		// Status might be "processing" or "completed" depending on queue
		status := stopResp.Session.SessionStatus
		assert.True(t, status == generated.SessionSessionStatusProcessing || status == generated.SessionSessionStatusCompleted,
			"status should be processing or completed, got %s", status)
	})

	// Step 5: Verify session in list
	t.Run("Step 5: Session appears in list", func(t *testing.T) {
		req := generated.RoutesListRequestObject{
			Params: generated.RoutesListParams{
				Authorization: authHeader,
			},
		}

		resp, err := sessionController.RoutesList(ctx, req)
		require.NoError(t, err)

		listResp, ok := resp.(generated.RoutesList200JSONResponse)
		require.True(t, ok, "expected 200 response")
		assert.GreaterOrEqual(t, len(listResp.Sessions), 1)

		// Find our session
		var found bool
		for _, s := range listResp.Sessions {
			if s.Id == sessionID {
				found = true
				break
			}
		}
		assert.True(t, found, "session should be in list")
	})
}

// TestSessionFlow_UpdateAndDelete tests session update and delete operations.
func TestSessionFlow_UpdateAndDelete(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	sessionService := service.NewSessionService(client, nil)
	sessionController := controller.NewSessionController(sessionService)

	ctx := context.Background()
	authHeader := getAuthHeader()

	// Setup: Create session
	var sessionID string

	t.Run("Setup: Create session", func(t *testing.T) {
		startReq := generated.RoutesStartRequestObject{
			Params: generated.RoutesStartParams{Authorization: authHeader},
		}
		startResp, err := sessionController.RoutesStart(ctx, startReq)
		require.NoError(t, err)
		sessionID = startResp.(generated.RoutesStart201JSONResponse).Session.Id
	})

	// Step 1: Update session
	t.Run("Step 1: Update session", func(t *testing.T) {
		newTitle := "Updated Title"
		newDesc := "Session description"
		req := generated.RoutesUpdateRequestObject{
			Params: generated.RoutesUpdateParams{Authorization: authHeader},
			Id:     sessionID,
			Body: &generated.RoutesUpdateJSONRequestBody{
				Title:       &newTitle,
				Description: &newDesc,
			},
		}

		resp, err := sessionController.RoutesUpdate(ctx, req)
		require.NoError(t, err)

		updateResp, ok := resp.(generated.RoutesUpdate200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, newTitle, *updateResp.Session.Title)
		assert.Equal(t, newDesc, *updateResp.Session.Description)
	})

	// Step 2: Get session to verify
	t.Run("Step 2: Get session", func(t *testing.T) {
		req := generated.RoutesGetRequestObject{
			Params: generated.RoutesGetParams{Authorization: authHeader},
			Id:     sessionID,
		}

		resp, err := sessionController.RoutesGet(ctx, req)
		require.NoError(t, err)

		getResp, ok := resp.(generated.RoutesGet200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, "Updated Title", *getResp.Session.Title)
	})

	// Step 3: Delete session
	t.Run("Step 3: Delete session", func(t *testing.T) {
		req := generated.RoutesDeleteRequestObject{
			Params: generated.RoutesDeleteParams{Authorization: authHeader},
			Id:     sessionID,
		}

		resp, err := sessionController.RoutesDelete(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesDelete204Response)
		require.True(t, ok, "expected 204 response")
	})

	// Step 4: Verify session is deleted
	t.Run("Step 4: Session not found", func(t *testing.T) {
		req := generated.RoutesGetRequestObject{
			Params: generated.RoutesGetParams{Authorization: authHeader},
			Id:     sessionID,
		}

		resp, err := sessionController.RoutesGet(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesGet404JSONResponse)
		require.True(t, ok, "deleted session should return 404")
	})
}

// TestSessionFlow_MultipleActiveSessions tests handling multiple sessions.
func TestSessionFlow_MultipleActiveSessions(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	sessionService := service.NewSessionService(client, nil)
	sessionController := controller.NewSessionController(sessionService)

	ctx := context.Background()
	authHeader := getAuthHeader()

	// Create first session
	startReq1 := generated.RoutesStartRequestObject{
		Params: generated.RoutesStartParams{Authorization: authHeader},
	}
	resp1, err := sessionController.RoutesStart(ctx, startReq1)
	require.NoError(t, err)
	session1ID := resp1.(generated.RoutesStart201JSONResponse).Session.Id

	// Small delay to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	// Create second session
	startReq2 := generated.RoutesStartRequestObject{
		Params: generated.RoutesStartParams{Authorization: authHeader},
	}
	resp2, err := sessionController.RoutesStart(ctx, startReq2)
	require.NoError(t, err)
	session2ID := resp2.(generated.RoutesStart201JSONResponse).Session.Id

	// Both sessions should be in list
	t.Run("Both sessions in list", func(t *testing.T) {
		listReq := generated.RoutesListRequestObject{
			Params: generated.RoutesListParams{Authorization: authHeader},
		}

		listResp, err := sessionController.RoutesList(ctx, listReq)
		require.NoError(t, err)

		sessions := listResp.(generated.RoutesList200JSONResponse).Sessions
		assert.GreaterOrEqual(t, len(sessions), 2)

		// Find both sessions
		foundIDs := make(map[string]bool)
		for _, s := range sessions {
			foundIDs[s.Id] = true
		}
		assert.True(t, foundIDs[session1ID], "session 1 should be in list")
		assert.True(t, foundIDs[session2ID], "session 2 should be in list")
	})

	// Both sessions can be operated independently
	t.Run("Sessions operate independently", func(t *testing.T) {
		// Pause session 1
		pauseReq := generated.RoutesPauseRequestObject{
			Params: generated.RoutesPauseParams{Authorization: authHeader},
			Id:     session1ID,
		}
		pauseResp, err := sessionController.RoutesPause(ctx, pauseReq)
		require.NoError(t, err)
		assert.Equal(t, generated.SessionSessionStatusPaused, pauseResp.(generated.RoutesPause200JSONResponse).Session.SessionStatus)

		// Session 2 should still be recording
		getReq := generated.RoutesGetRequestObject{
			Params: generated.RoutesGetParams{Authorization: authHeader},
			Id:     session2ID,
		}
		getResp, err := sessionController.RoutesGet(ctx, getReq)
		require.NoError(t, err)
		assert.Equal(t, generated.SessionSessionStatusRecording, getResp.(generated.RoutesGet200JSONResponse).Session.SessionStatus)
	})
}
