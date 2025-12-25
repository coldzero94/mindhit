package controller

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mindhit/api/internal/generated"
	"github.com/mindhit/api/internal/service"
	"github.com/mindhit/api/internal/testutil"
)

func TestAuthController_RoutesSignup(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	authService := service.NewAuthService(client)
	jwtService := service.NewJWTService("test-secret")
	controller := NewAuthController(authService, jwtService)

	ctx := context.Background()

	t.Run("successful signup", func(t *testing.T) {
		req := generated.RoutesSignupRequestObject{
			Body: &generated.RoutesSignupJSONRequestBody{
				Email:    "signup@test.com",
				Password: "password123",
			},
		}

		resp, err := controller.RoutesSignup(ctx, req)
		require.NoError(t, err)

		successResp, ok := resp.(generated.RoutesSignup201JSONResponse)
		require.True(t, ok, "expected 201 response")
		assert.NotEmpty(t, successResp.Token)
		assert.Equal(t, "signup@test.com", successResp.User.Email)
		assert.NotEmpty(t, successResp.User.Id)
	})

	t.Run("duplicate email returns 409", func(t *testing.T) {
		// First signup
		req := generated.RoutesSignupRequestObject{
			Body: &generated.RoutesSignupJSONRequestBody{
				Email:    "duplicate@test.com",
				Password: "password123",
			},
		}
		_, err := controller.RoutesSignup(ctx, req)
		require.NoError(t, err)

		// Second signup with same email
		resp, err := controller.RoutesSignup(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesSignup409JSONResponse)
		assert.True(t, ok, "expected 409 response")
	})

	t.Run("nil body returns 400", func(t *testing.T) {
		req := generated.RoutesSignupRequestObject{
			Body: nil,
		}

		resp, err := controller.RoutesSignup(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesSignup400JSONResponse)
		assert.True(t, ok, "expected 400 response")
	})
}

func TestAuthController_RoutesLogin(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	authService := service.NewAuthService(client)
	jwtService := service.NewJWTService("test-secret")
	controller := NewAuthController(authService, jwtService)

	ctx := context.Background()

	// Create test user
	_, err := authService.Signup(ctx, "login@test.com", "password123")
	require.NoError(t, err)

	t.Run("successful login", func(t *testing.T) {
		req := generated.RoutesLoginRequestObject{
			Body: &generated.RoutesLoginJSONRequestBody{
				Email:    "login@test.com",
				Password: "password123",
			},
		}

		resp, err := controller.RoutesLogin(ctx, req)
		require.NoError(t, err)

		successResp, ok := resp.(generated.RoutesLogin200JSONResponse)
		require.True(t, ok, "expected 200 response")
		assert.NotEmpty(t, successResp.Token)
		assert.Equal(t, "login@test.com", successResp.User.Email)
	})

	t.Run("wrong password returns 401", func(t *testing.T) {
		req := generated.RoutesLoginRequestObject{
			Body: &generated.RoutesLoginJSONRequestBody{
				Email:    "login@test.com",
				Password: "wrongpassword",
			},
		}

		resp, err := controller.RoutesLogin(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesLogin401JSONResponse)
		assert.True(t, ok, "expected 401 response")
	})

	t.Run("non-existent user returns 401", func(t *testing.T) {
		req := generated.RoutesLoginRequestObject{
			Body: &generated.RoutesLoginJSONRequestBody{
				Email:    "nonexistent@test.com",
				Password: "password123",
			},
		}

		resp, err := controller.RoutesLogin(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesLogin401JSONResponse)
		assert.True(t, ok, "expected 401 response")
	})

	t.Run("nil body returns 401", func(t *testing.T) {
		req := generated.RoutesLoginRequestObject{
			Body: nil,
		}

		resp, err := controller.RoutesLogin(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesLogin401JSONResponse)
		assert.True(t, ok, "expected 401 response")
	})
}

func TestAuthController_RoutesRefresh(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	authService := service.NewAuthService(client)
	jwtService := service.NewJWTService("test-secret")
	controller := NewAuthController(authService, jwtService)

	ctx := context.Background()

	// Create test user and get tokens
	user, err := authService.Signup(ctx, "refresh@test.com", "password123")
	require.NoError(t, err)

	tokenPair, err := jwtService.GenerateTokenPair(user.ID)
	require.NoError(t, err)

	t.Run("successful refresh", func(t *testing.T) {
		req := generated.RoutesRefreshRequestObject{
			Params: generated.RoutesRefreshParams{
				Authorization: "Bearer " + tokenPair.RefreshToken,
			},
		}

		resp, err := controller.RoutesRefresh(ctx, req)
		require.NoError(t, err)

		successResp, ok := resp.(generated.RoutesRefresh200JSONResponse)
		require.True(t, ok, "expected 200 response")
		assert.NotEmpty(t, successResp.Token)
		// Token should be a valid access token
		claims, err := jwtService.ValidateAccessToken(successResp.Token)
		require.NoError(t, err)
		assert.Equal(t, user.ID, claims.UserID)
	})

	t.Run("missing authorization header returns 401", func(t *testing.T) {
		req := generated.RoutesRefreshRequestObject{
			Params: generated.RoutesRefreshParams{
				Authorization: "",
			},
		}

		resp, err := controller.RoutesRefresh(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesRefresh401JSONResponse)
		assert.True(t, ok, "expected 401 response")
	})

	t.Run("invalid token format returns 401", func(t *testing.T) {
		req := generated.RoutesRefreshRequestObject{
			Params: generated.RoutesRefreshParams{
				Authorization: "InvalidFormat",
			},
		}

		resp, err := controller.RoutesRefresh(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesRefresh401JSONResponse)
		assert.True(t, ok, "expected 401 response")
	})

	t.Run("access token instead of refresh token returns 401", func(t *testing.T) {
		req := generated.RoutesRefreshRequestObject{
			Params: generated.RoutesRefreshParams{
				Authorization: "Bearer " + tokenPair.AccessToken,
			},
		}

		resp, err := controller.RoutesRefresh(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesRefresh401JSONResponse)
		assert.True(t, ok, "expected 401 response when using access token for refresh")
	})
}

func TestAuthController_RoutesMe(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	authService := service.NewAuthService(client)
	jwtService := service.NewJWTService("test-secret")
	controller := NewAuthController(authService, jwtService)

	ctx := context.Background()

	// Create test user and get tokens
	user, err := authService.Signup(ctx, "me@test.com", "password123")
	require.NoError(t, err)

	tokenPair, err := jwtService.GenerateTokenPair(user.ID)
	require.NoError(t, err)

	t.Run("successful get user info", func(t *testing.T) {
		req := generated.RoutesMeRequestObject{
			Params: generated.RoutesMeParams{
				Authorization: "Bearer " + tokenPair.AccessToken,
			},
		}

		resp, err := controller.RoutesMe(ctx, req)
		require.NoError(t, err)

		successResp, ok := resp.(generated.RoutesMe200JSONResponse)
		require.True(t, ok, "expected 200 response")
		assert.Equal(t, "me@test.com", successResp.User.Email)
		assert.Equal(t, user.ID.String(), successResp.User.Id)
	})

	t.Run("missing authorization header returns 401", func(t *testing.T) {
		req := generated.RoutesMeRequestObject{
			Params: generated.RoutesMeParams{
				Authorization: "",
			},
		}

		resp, err := controller.RoutesMe(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesMe401JSONResponse)
		assert.True(t, ok, "expected 401 response")
	})

	t.Run("invalid token returns 401", func(t *testing.T) {
		req := generated.RoutesMeRequestObject{
			Params: generated.RoutesMeParams{
				Authorization: "Bearer invalid-token",
			},
		}

		resp, err := controller.RoutesMe(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesMe401JSONResponse)
		assert.True(t, ok, "expected 401 response")
	})

	t.Run("refresh token instead of access token returns 401", func(t *testing.T) {
		req := generated.RoutesMeRequestObject{
			Params: generated.RoutesMeParams{
				Authorization: "Bearer " + tokenPair.RefreshToken,
			},
		}

		resp, err := controller.RoutesMe(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesMe401JSONResponse)
		assert.True(t, ok, "expected 401 response when using refresh token")
	})
}

func TestAuthController_RoutesLogout(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	authService := service.NewAuthService(client)
	jwtService := service.NewJWTService("test-secret")
	controller := NewAuthController(authService, jwtService)

	ctx := context.Background()

	// Create test user and get tokens
	user, err := authService.Signup(ctx, "logout@test.com", "password123")
	require.NoError(t, err)

	tokenPair, err := jwtService.GenerateTokenPair(user.ID)
	require.NoError(t, err)

	t.Run("successful logout", func(t *testing.T) {
		req := generated.RoutesLogoutRequestObject{
			Params: generated.RoutesLogoutParams{
				Authorization: "Bearer " + tokenPair.AccessToken,
			},
		}

		resp, err := controller.RoutesLogout(ctx, req)
		require.NoError(t, err)

		successResp, ok := resp.(generated.RoutesLogout200JSONResponse)
		require.True(t, ok, "expected 200 response")
		assert.Equal(t, "successfully logged out", successResp.Message)
	})

	t.Run("missing authorization header returns 401", func(t *testing.T) {
		req := generated.RoutesLogoutRequestObject{
			Params: generated.RoutesLogoutParams{
				Authorization: "",
			},
		}

		resp, err := controller.RoutesLogout(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesLogout401JSONResponse)
		assert.True(t, ok, "expected 401 response")
	})

	t.Run("invalid token returns 401", func(t *testing.T) {
		req := generated.RoutesLogoutRequestObject{
			Params: generated.RoutesLogoutParams{
				Authorization: "Bearer invalid-token",
			},
		}

		resp, err := controller.RoutesLogout(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesLogout401JSONResponse)
		assert.True(t, ok, "expected 401 response")
	})
}

func TestAuthController_RoutesForgotPassword(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	authService := service.NewAuthService(client)
	jwtService := service.NewJWTService("test-secret")
	controller := NewAuthController(authService, jwtService)

	ctx := context.Background()

	// Create test user
	_, err := authService.Signup(ctx, "forgot@test.com", "password123")
	require.NoError(t, err)

	t.Run("successful forgot password request", func(t *testing.T) {
		req := generated.RoutesForgotPasswordRequestObject{
			Body: &generated.RoutesForgotPasswordJSONRequestBody{
				Email: "forgot@test.com",
			},
		}

		resp, err := controller.RoutesForgotPassword(ctx, req)
		require.NoError(t, err)

		successResp, ok := resp.(generated.RoutesForgotPassword200JSONResponse)
		require.True(t, ok, "expected 200 response")
		assert.Contains(t, successResp.Message, "password reset link")
	})

	t.Run("non-existent email still returns 200 (security)", func(t *testing.T) {
		req := generated.RoutesForgotPasswordRequestObject{
			Body: &generated.RoutesForgotPasswordJSONRequestBody{
				Email: "nonexistent@test.com",
			},
		}

		resp, err := controller.RoutesForgotPassword(ctx, req)
		require.NoError(t, err)

		successResp, ok := resp.(generated.RoutesForgotPassword200JSONResponse)
		require.True(t, ok, "expected 200 response even for non-existent email")
		assert.Contains(t, successResp.Message, "password reset link")
	})

	t.Run("nil body returns 400", func(t *testing.T) {
		req := generated.RoutesForgotPasswordRequestObject{
			Body: nil,
		}

		resp, err := controller.RoutesForgotPassword(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesForgotPassword400JSONResponse)
		assert.True(t, ok, "expected 400 response")
	})
}

func TestAuthController_RoutesResetPassword(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	authService := service.NewAuthService(client)
	jwtService := service.NewJWTService("test-secret")
	controller := NewAuthController(authService, jwtService)

	ctx := context.Background()

	t.Run("successful password reset", func(t *testing.T) {
		// Create test user and generate reset token for this test
		_, err := authService.Signup(ctx, "reset1@test.com", "oldpassword123")
		require.NoError(t, err)

		token, err := authService.RequestPasswordReset(ctx, "reset1@test.com")
		require.NoError(t, err)
		require.NotEmpty(t, token)

		req := generated.RoutesResetPasswordRequestObject{
			Body: &generated.RoutesResetPasswordJSONRequestBody{
				Token:       token,
				NewPassword: "newpassword123",
			},
		}

		resp, err := controller.RoutesResetPassword(ctx, req)
		require.NoError(t, err)

		successResp, ok := resp.(generated.RoutesResetPassword200JSONResponse)
		require.True(t, ok, "expected 200 response")
		assert.Contains(t, successResp.Message, "successfully")

		// Verify login with new password works
		user, err := authService.Login(ctx, "reset1@test.com", "newpassword123")
		require.NoError(t, err)
		assert.Equal(t, "reset1@test.com", user.Email)
	})

	t.Run("invalid token returns 400", func(t *testing.T) {
		req := generated.RoutesResetPasswordRequestObject{
			Body: &generated.RoutesResetPasswordJSONRequestBody{
				Token:       "invalid-token",
				NewPassword: "newpassword123",
			},
		}

		resp, err := controller.RoutesResetPassword(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesResetPassword400JSONResponse)
		assert.True(t, ok, "expected 400 response")
	})

	t.Run("already used token returns 400", func(t *testing.T) {
		// Create another user with a token
		_, err := authService.Signup(ctx, "reset2@test.com", "password123")
		require.NoError(t, err)

		token2, err := authService.RequestPasswordReset(ctx, "reset2@test.com")
		require.NoError(t, err)

		// Use the token
		err = authService.ResetPassword(ctx, token2, "newpassword")
		require.NoError(t, err)

		// Try to use again
		req := generated.RoutesResetPasswordRequestObject{
			Body: &generated.RoutesResetPasswordJSONRequestBody{
				Token:       token2,
				NewPassword: "anotherpassword",
			},
		}

		resp, err := controller.RoutesResetPassword(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesResetPassword400JSONResponse)
		assert.True(t, ok, "expected 400 response for used token")
	})

	t.Run("nil body returns 400", func(t *testing.T) {
		req := generated.RoutesResetPasswordRequestObject{
			Body: nil,
		}

		resp, err := controller.RoutesResetPassword(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesResetPassword400JSONResponse)
		assert.True(t, ok, "expected 400 response")
	})
}
