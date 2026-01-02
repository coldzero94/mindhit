package controller

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mindhit/api/internal/generated"
	"github.com/mindhit/api/internal/service"
	"github.com/mindhit/api/internal/testutil"
)

// uniqueEmail generates a unique email for each test
func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s-%s@test.com", prefix, uuid.New().String()[:8])
}

func setupSessionControllerTest(t *testing.T) (*SessionController, *service.AuthService, *service.JWTService, func()) {
	client := testutil.SetupTestDB(t)
	authService := service.NewAuthService(client)
	jwtService := service.NewJWTService("test-secret")
	sessionService := service.NewSessionService(client, nil) // nil queue client for tests
	controller := NewSessionController(sessionService, jwtService)

	cleanup := func() {
		testutil.CleanupTestDB(t, client)
	}

	return controller, authService, jwtService, cleanup
}

func createUserAndGetToken(t *testing.T, authService *service.AuthService, jwtService *service.JWTService, email string) string {
	ctx := context.Background()
	user, err := authService.Signup(ctx, email, "password123")
	require.NoError(t, err)

	tokenPair, err := jwtService.GenerateTokenPair(user.ID)
	require.NoError(t, err)

	return "Bearer " + tokenPair.AccessToken
}

// ==================== Start Tests ====================

func TestSessionController_RoutesStart(t *testing.T) {
	controller, authService, jwtService, cleanup := setupSessionControllerTest(t)
	defer cleanup()

	ctx := context.Background()
	token := createUserAndGetToken(t, authService, jwtService, uniqueEmail("test"))

	t.Run("successful start", func(t *testing.T) {
		req := generated.RoutesStartRequestObject{
			Params: generated.RoutesStartParams{
				Authorization: token,
			},
		}

		resp, err := controller.RoutesStart(ctx, req)
		require.NoError(t, err)

		successResp, ok := resp.(generated.RoutesStart201JSONResponse)
		require.True(t, ok, "expected 201 response")
		assert.NotEmpty(t, successResp.Session.Id)
		assert.Equal(t, generated.SessionSessionStatusRecording, successResp.Session.SessionStatus)
	})

	t.Run("missing authorization returns 401", func(t *testing.T) {
		req := generated.RoutesStartRequestObject{
			Params: generated.RoutesStartParams{
				Authorization: "",
			},
		}

		resp, err := controller.RoutesStart(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesStart401JSONResponse)
		assert.True(t, ok, "expected 401 response")
	})

	t.Run("invalid token returns 401", func(t *testing.T) {
		req := generated.RoutesStartRequestObject{
			Params: generated.RoutesStartParams{
				Authorization: "Bearer invalid-token",
			},
		}

		resp, err := controller.RoutesStart(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesStart401JSONResponse)
		assert.True(t, ok, "expected 401 response")
	})
}

// ==================== List Tests ====================

func TestSessionController_RoutesList(t *testing.T) {
	controller, authService, jwtService, cleanup := setupSessionControllerTest(t)
	defer cleanup()

	ctx := context.Background()
	token := createUserAndGetToken(t, authService, jwtService, uniqueEmail("test"))

	t.Run("list empty sessions", func(t *testing.T) {
		limit := int32(20)
		offset := int32(0)
		req := generated.RoutesListRequestObject{
			Params: generated.RoutesListParams{
				Authorization: token,
				Limit:         &limit,
				Offset:        &offset,
			},
		}

		resp, err := controller.RoutesList(ctx, req)
		require.NoError(t, err)

		successResp, ok := resp.(generated.RoutesList200JSONResponse)
		require.True(t, ok, "expected 200 response")
		assert.Empty(t, successResp.Sessions)
	})

	t.Run("list sessions after creating some", func(t *testing.T) {
		// Create some sessions first
		startReq := generated.RoutesStartRequestObject{
			Params: generated.RoutesStartParams{
				Authorization: token,
			},
		}
		_, _ = controller.RoutesStart(ctx, startReq)
		_, _ = controller.RoutesStart(ctx, startReq)

		limit := int32(20)
		offset := int32(0)
		req := generated.RoutesListRequestObject{
			Params: generated.RoutesListParams{
				Authorization: token,
				Limit:         &limit,
				Offset:        &offset,
			},
		}

		resp, err := controller.RoutesList(ctx, req)
		require.NoError(t, err)

		successResp, ok := resp.(generated.RoutesList200JSONResponse)
		require.True(t, ok, "expected 200 response")
		assert.Len(t, successResp.Sessions, 2)
	})

	t.Run("missing authorization returns 401", func(t *testing.T) {
		req := generated.RoutesListRequestObject{
			Params: generated.RoutesListParams{
				Authorization: "",
			},
		}

		resp, err := controller.RoutesList(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesList401JSONResponse)
		assert.True(t, ok, "expected 401 response")
	})
}

// ==================== Get Tests ====================

func TestSessionController_RoutesGet(t *testing.T) {
	controller, authService, jwtService, cleanup := setupSessionControllerTest(t)
	defer cleanup()

	ctx := context.Background()
	token := createUserAndGetToken(t, authService, jwtService, uniqueEmail("test"))

	// Create a session first
	startReq := generated.RoutesStartRequestObject{
		Params: generated.RoutesStartParams{
			Authorization: token,
		},
	}
	startResp, err := controller.RoutesStart(ctx, startReq)
	require.NoError(t, err)
	sessionID := startResp.(generated.RoutesStart201JSONResponse).Session.Id

	t.Run("successful get", func(t *testing.T) {
		req := generated.RoutesGetRequestObject{
			Id: sessionID,
			Params: generated.RoutesGetParams{
				Authorization: token,
			},
		}

		resp, err := controller.RoutesGet(ctx, req)
		require.NoError(t, err)

		successResp, ok := resp.(generated.RoutesGet200JSONResponse)
		require.True(t, ok, "expected 200 response")
		assert.Equal(t, sessionID, successResp.Session.Id)
	})

	t.Run("invalid uuid returns 404", func(t *testing.T) {
		req := generated.RoutesGetRequestObject{
			Id: "invalid-uuid",
			Params: generated.RoutesGetParams{
				Authorization: token,
			},
		}

		resp, err := controller.RoutesGet(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesGet404JSONResponse)
		assert.True(t, ok, "expected 404 response")
	})

	t.Run("non-existent session returns 404", func(t *testing.T) {
		req := generated.RoutesGetRequestObject{
			Id: uuid.New().String(),
			Params: generated.RoutesGetParams{
				Authorization: token,
			},
		}

		resp, err := controller.RoutesGet(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesGet404JSONResponse)
		assert.True(t, ok, "expected 404 response")
	})

	t.Run("other user's session returns 403", func(t *testing.T) {
		otherToken := createUserAndGetToken(t, authService, jwtService, uniqueEmail("other"))

		req := generated.RoutesGetRequestObject{
			Id: sessionID,
			Params: generated.RoutesGetParams{
				Authorization: otherToken,
			},
		}

		resp, err := controller.RoutesGet(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesGet403JSONResponse)
		assert.True(t, ok, "expected 403 response")
	})

	t.Run("missing authorization returns 401", func(t *testing.T) {
		req := generated.RoutesGetRequestObject{
			Id: sessionID,
			Params: generated.RoutesGetParams{
				Authorization: "",
			},
		}

		resp, err := controller.RoutesGet(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesGet401JSONResponse)
		assert.True(t, ok, "expected 401 response")
	})
}

// ==================== Pause Tests ====================

func TestSessionController_RoutesPause(t *testing.T) {
	controller, authService, jwtService, cleanup := setupSessionControllerTest(t)
	defer cleanup()

	ctx := context.Background()
	token := createUserAndGetToken(t, authService, jwtService, uniqueEmail("test"))

	// Create a session first
	startReq := generated.RoutesStartRequestObject{
		Params: generated.RoutesStartParams{
			Authorization: token,
		},
	}
	startResp, err := controller.RoutesStart(ctx, startReq)
	require.NoError(t, err)
	sessionID := startResp.(generated.RoutesStart201JSONResponse).Session.Id

	t.Run("successful pause", func(t *testing.T) {
		req := generated.RoutesPauseRequestObject{
			Id: sessionID,
			Params: generated.RoutesPauseParams{
				Authorization: token,
			},
		}

		resp, err := controller.RoutesPause(ctx, req)
		require.NoError(t, err)

		successResp, ok := resp.(generated.RoutesPause200JSONResponse)
		require.True(t, ok, "expected 200 response")
		assert.Equal(t, generated.SessionSessionStatusPaused, successResp.Session.SessionStatus)
	})

	t.Run("pause already paused returns 400", func(t *testing.T) {
		req := generated.RoutesPauseRequestObject{
			Id: sessionID,
			Params: generated.RoutesPauseParams{
				Authorization: token,
			},
		}

		resp, err := controller.RoutesPause(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesPause400JSONResponse)
		assert.True(t, ok, "expected 400 response for already paused session")
	})

	t.Run("missing authorization returns 401", func(t *testing.T) {
		req := generated.RoutesPauseRequestObject{
			Id: sessionID,
			Params: generated.RoutesPauseParams{
				Authorization: "",
			},
		}

		resp, err := controller.RoutesPause(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesPause401JSONResponse)
		assert.True(t, ok, "expected 401 response")
	})
}

// ==================== Resume Tests ====================

func TestSessionController_RoutesResume(t *testing.T) {
	controller, authService, jwtService, cleanup := setupSessionControllerTest(t)
	defer cleanup()

	ctx := context.Background()
	token := createUserAndGetToken(t, authService, jwtService, uniqueEmail("test"))

	// Create and pause a session first
	startReq := generated.RoutesStartRequestObject{
		Params: generated.RoutesStartParams{
			Authorization: token,
		},
	}
	startResp, err := controller.RoutesStart(ctx, startReq)
	require.NoError(t, err)
	sessionID := startResp.(generated.RoutesStart201JSONResponse).Session.Id

	pauseReq := generated.RoutesPauseRequestObject{
		Id: sessionID,
		Params: generated.RoutesPauseParams{
			Authorization: token,
		},
	}
	_, err = controller.RoutesPause(ctx, pauseReq)
	require.NoError(t, err)

	t.Run("successful resume", func(t *testing.T) {
		req := generated.RoutesResumeRequestObject{
			Id: sessionID,
			Params: generated.RoutesResumeParams{
				Authorization: token,
			},
		}

		resp, err := controller.RoutesResume(ctx, req)
		require.NoError(t, err)

		successResp, ok := resp.(generated.RoutesResume200JSONResponse)
		require.True(t, ok, "expected 200 response")
		assert.Equal(t, generated.SessionSessionStatusRecording, successResp.Session.SessionStatus)
	})

	t.Run("resume recording session returns 400", func(t *testing.T) {
		req := generated.RoutesResumeRequestObject{
			Id: sessionID,
			Params: generated.RoutesResumeParams{
				Authorization: token,
			},
		}

		resp, err := controller.RoutesResume(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesResume400JSONResponse)
		assert.True(t, ok, "expected 400 response for already recording session")
	})

	t.Run("missing authorization returns 401", func(t *testing.T) {
		req := generated.RoutesResumeRequestObject{
			Id: sessionID,
			Params: generated.RoutesResumeParams{
				Authorization: "",
			},
		}

		resp, err := controller.RoutesResume(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesResume401JSONResponse)
		assert.True(t, ok, "expected 401 response")
	})
}

// ==================== Stop Tests ====================

func TestSessionController_RoutesStop(t *testing.T) {
	controller, authService, jwtService, cleanup := setupSessionControllerTest(t)
	defer cleanup()

	ctx := context.Background()
	token := createUserAndGetToken(t, authService, jwtService, uniqueEmail("test"))

	// Create a session first
	startReq := generated.RoutesStartRequestObject{
		Params: generated.RoutesStartParams{
			Authorization: token,
		},
	}
	startResp, err := controller.RoutesStart(ctx, startReq)
	require.NoError(t, err)
	sessionID := startResp.(generated.RoutesStart201JSONResponse).Session.Id

	t.Run("successful stop", func(t *testing.T) {
		req := generated.RoutesStopRequestObject{
			Id: sessionID,
			Params: generated.RoutesStopParams{
				Authorization: token,
			},
		}

		resp, err := controller.RoutesStop(ctx, req)
		require.NoError(t, err)

		successResp, ok := resp.(generated.RoutesStop200JSONResponse)
		require.True(t, ok, "expected 200 response")
		assert.Equal(t, generated.SessionSessionStatusProcessing, successResp.Session.SessionStatus)
		assert.NotNil(t, successResp.Session.EndedAt)
	})

	t.Run("stop already stopped returns 400", func(t *testing.T) {
		req := generated.RoutesStopRequestObject{
			Id: sessionID,
			Params: generated.RoutesStopParams{
				Authorization: token,
			},
		}

		resp, err := controller.RoutesStop(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesStop400JSONResponse)
		assert.True(t, ok, "expected 400 response for already stopped session")
	})

	t.Run("missing authorization returns 401", func(t *testing.T) {
		req := generated.RoutesStopRequestObject{
			Id: sessionID,
			Params: generated.RoutesStopParams{
				Authorization: "",
			},
		}

		resp, err := controller.RoutesStop(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesStop401JSONResponse)
		assert.True(t, ok, "expected 401 response")
	})
}

// ==================== Update Tests ====================

func TestSessionController_RoutesUpdate(t *testing.T) {
	controller, authService, jwtService, cleanup := setupSessionControllerTest(t)
	defer cleanup()

	ctx := context.Background()
	token := createUserAndGetToken(t, authService, jwtService, uniqueEmail("test"))

	// Create a session first
	startReq := generated.RoutesStartRequestObject{
		Params: generated.RoutesStartParams{
			Authorization: token,
		},
	}
	startResp, err := controller.RoutesStart(ctx, startReq)
	require.NoError(t, err)
	sessionID := startResp.(generated.RoutesStart201JSONResponse).Session.Id

	t.Run("successful update", func(t *testing.T) {
		title := "My Research Session"
		description := "Researching AI topics"
		req := generated.RoutesUpdateRequestObject{
			Id: sessionID,
			Params: generated.RoutesUpdateParams{
				Authorization: token,
			},
			Body: &generated.RoutesUpdateJSONRequestBody{
				Title:       &title,
				Description: &description,
			},
		}

		resp, err := controller.RoutesUpdate(ctx, req)
		require.NoError(t, err)

		successResp, ok := resp.(generated.RoutesUpdate200JSONResponse)
		require.True(t, ok, "expected 200 response")
		assert.Equal(t, title, *successResp.Session.Title)
		assert.Equal(t, description, *successResp.Session.Description)
	})

	t.Run("partial update (title only)", func(t *testing.T) {
		title := "Updated Title"
		req := generated.RoutesUpdateRequestObject{
			Id: sessionID,
			Params: generated.RoutesUpdateParams{
				Authorization: token,
			},
			Body: &generated.RoutesUpdateJSONRequestBody{
				Title: &title,
			},
		}

		resp, err := controller.RoutesUpdate(ctx, req)
		require.NoError(t, err)

		successResp, ok := resp.(generated.RoutesUpdate200JSONResponse)
		require.True(t, ok, "expected 200 response")
		assert.Equal(t, title, *successResp.Session.Title)
	})

	t.Run("other user cannot update returns 403", func(t *testing.T) {
		otherToken := createUserAndGetToken(t, authService, jwtService, uniqueEmail("other"))

		title := "Hacked"
		req := generated.RoutesUpdateRequestObject{
			Id: sessionID,
			Params: generated.RoutesUpdateParams{
				Authorization: otherToken,
			},
			Body: &generated.RoutesUpdateJSONRequestBody{
				Title: &title,
			},
		}

		resp, err := controller.RoutesUpdate(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesUpdate403JSONResponse)
		assert.True(t, ok, "expected 403 response")
	})

	t.Run("missing authorization returns 401", func(t *testing.T) {
		title := "Test"
		req := generated.RoutesUpdateRequestObject{
			Id: sessionID,
			Params: generated.RoutesUpdateParams{
				Authorization: "",
			},
			Body: &generated.RoutesUpdateJSONRequestBody{
				Title: &title,
			},
		}

		resp, err := controller.RoutesUpdate(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesUpdate401JSONResponse)
		assert.True(t, ok, "expected 401 response")
	})
}

// ==================== Delete Tests ====================

func TestSessionController_RoutesDelete(t *testing.T) {
	controller, authService, jwtService, cleanup := setupSessionControllerTest(t)
	defer cleanup()

	ctx := context.Background()
	token := createUserAndGetToken(t, authService, jwtService, uniqueEmail("test"))

	// Create a session first
	startReq := generated.RoutesStartRequestObject{
		Params: generated.RoutesStartParams{
			Authorization: token,
		},
	}
	startResp, err := controller.RoutesStart(ctx, startReq)
	require.NoError(t, err)
	sessionID := startResp.(generated.RoutesStart201JSONResponse).Session.Id

	t.Run("successful delete", func(t *testing.T) {
		req := generated.RoutesDeleteRequestObject{
			Id: sessionID,
			Params: generated.RoutesDeleteParams{
				Authorization: token,
			},
		}

		resp, err := controller.RoutesDelete(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesDelete204Response)
		assert.True(t, ok, "expected 204 response")

		// Verify session is deleted
		getReq := generated.RoutesGetRequestObject{
			Id: sessionID,
			Params: generated.RoutesGetParams{
				Authorization: token,
			},
		}
		getResp, err := controller.RoutesGet(ctx, getReq)
		require.NoError(t, err)
		_, ok = getResp.(generated.RoutesGet404JSONResponse)
		assert.True(t, ok, "expected 404 after deletion")
	})

	t.Run("delete non-existent session returns 404", func(t *testing.T) {
		req := generated.RoutesDeleteRequestObject{
			Id: uuid.New().String(),
			Params: generated.RoutesDeleteParams{
				Authorization: token,
			},
		}

		resp, err := controller.RoutesDelete(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesDelete404JSONResponse)
		assert.True(t, ok, "expected 404 response")
	})

	t.Run("other user cannot delete returns 403", func(t *testing.T) {
		// Create another session
		startResp, err := controller.RoutesStart(ctx, startReq)
		require.NoError(t, err)
		sessionID := startResp.(generated.RoutesStart201JSONResponse).Session.Id

		otherToken := createUserAndGetToken(t, authService, jwtService, uniqueEmail("other"))

		req := generated.RoutesDeleteRequestObject{
			Id: sessionID,
			Params: generated.RoutesDeleteParams{
				Authorization: otherToken,
			},
		}

		resp, err := controller.RoutesDelete(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesDelete403JSONResponse)
		assert.True(t, ok, "expected 403 response")
	})

	t.Run("missing authorization returns 401", func(t *testing.T) {
		req := generated.RoutesDeleteRequestObject{
			Id: uuid.New().String(),
			Params: generated.RoutesDeleteParams{
				Authorization: "",
			},
		}

		resp, err := controller.RoutesDelete(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesDelete401JSONResponse)
		assert.True(t, ok, "expected 401 response")
	})
}
