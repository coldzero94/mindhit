//go:build integration

// Package integration contains end-to-end flow tests that test
// complete user journeys across multiple services.
package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mindhit/api/internal/controller"
	"github.com/mindhit/api/internal/generated"
	"github.com/mindhit/api/internal/service"
	"github.com/mindhit/api/internal/testutil"
)

// TestAuthFlow_SignupLoginRefreshLogout tests the complete authentication flow:
// 1. User signs up
// 2. User logs in
// 3. User refreshes token
// 4. User accesses protected endpoint (/me)
// 5. User logs out
func TestAuthFlow_SignupLoginRefreshLogout(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	authService := service.NewAuthService(client)
	jwtService := service.NewJWTService("test-secret")
	authController := controller.NewAuthController(authService, jwtService)

	ctx := context.Background()
	email := uniqueEmail("auth_flow")
	password := "securePassword123!"

	// Step 1: Signup
	t.Run("Step 1: Signup", func(t *testing.T) {
		req := generated.RoutesSignupRequestObject{
			Body: &generated.RoutesSignupJSONRequestBody{
				Email:    email,
				Password: password,
			},
		}

		resp, err := authController.RoutesSignup(ctx, req)
		require.NoError(t, err)

		signupResp, ok := resp.(generated.RoutesSignup201JSONResponse)
		require.True(t, ok, "expected 201 response on signup")
		assert.Equal(t, email, signupResp.User.Email)
		assert.NotEmpty(t, signupResp.Token)
	})

	// Step 2: Login and generate refresh token
	var accessToken, refreshToken string
	var userID string
	t.Run("Step 2: Login", func(t *testing.T) {
		req := generated.RoutesLoginRequestObject{
			Body: &generated.RoutesLoginJSONRequestBody{
				Email:    email,
				Password: password,
			},
		}

		resp, err := authController.RoutesLogin(ctx, req)
		require.NoError(t, err)

		loginResp, ok := resp.(generated.RoutesLogin200JSONResponse)
		require.True(t, ok, "expected 200 response on login")
		assert.Equal(t, email, loginResp.User.Email)

		accessToken = loginResp.Token
		userID = loginResp.User.Id
		assert.NotEmpty(t, accessToken)

		// Generate refresh token using JWT service for refresh test
		uid, err := parseUUID(userID)
		require.NoError(t, err)
		tokenPair, err := jwtService.GenerateTokenPair(uid)
		require.NoError(t, err)
		refreshToken = tokenPair.RefreshToken
	})

	// Step 3: Access protected endpoint (/me)
	t.Run("Step 3: Access /me endpoint", func(t *testing.T) {
		req := generated.RoutesMeRequestObject{
			Params: generated.RoutesMeParams{
				Authorization: "Bearer " + accessToken,
			},
		}

		resp, err := authController.RoutesMe(ctx, req)
		require.NoError(t, err)

		meResp, ok := resp.(generated.RoutesMe200JSONResponse)
		require.True(t, ok, "expected 200 response on /me")
		assert.Equal(t, email, meResp.User.Email)
	})

	// Step 4: Refresh token
	t.Run("Step 4: Refresh token", func(t *testing.T) {
		if refreshToken == "" {
			t.Skip("No refresh token available")
		}

		req := generated.RoutesRefreshRequestObject{
			Params: generated.RoutesRefreshParams{
				Authorization: "Bearer " + refreshToken,
			},
		}

		resp, err := authController.RoutesRefresh(ctx, req)
		require.NoError(t, err)

		refreshResp, ok := resp.(generated.RoutesRefresh200JSONResponse)
		require.True(t, ok, "expected 200 response on refresh")
		assert.NotEmpty(t, refreshResp.Token)

		// New token should be valid
		accessToken = refreshResp.Token
	})

	// Step 5: Verify new token works
	t.Run("Step 5: Verify new token works", func(t *testing.T) {
		req := generated.RoutesMeRequestObject{
			Params: generated.RoutesMeParams{
				Authorization: "Bearer " + accessToken,
			},
		}

		resp, err := authController.RoutesMe(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesMe200JSONResponse)
		require.True(t, ok, "expected 200 response with refreshed token")
	})

	// Step 6: Logout
	t.Run("Step 6: Logout", func(t *testing.T) {
		req := generated.RoutesLogoutRequestObject{
			Params: generated.RoutesLogoutParams{
				Authorization: "Bearer " + accessToken,
			},
		}

		resp, err := authController.RoutesLogout(ctx, req)
		require.NoError(t, err)

		logoutResp, ok := resp.(generated.RoutesLogout200JSONResponse)
		require.True(t, ok, "expected 200 response on logout")
		assert.Contains(t, logoutResp.Message, "logged out")
	})
}

// TestAuthFlow_PasswordReset tests the password reset flow:
// 1. User signs up
// 2. User requests password reset
// 3. User resets password with token
// 4. User logs in with new password
// 5. Old password no longer works
func TestAuthFlow_PasswordReset(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	authService := service.NewAuthService(client)
	jwtService := service.NewJWTService("test-secret")
	authController := controller.NewAuthController(authService, jwtService)

	ctx := context.Background()
	email := uniqueEmail("password_reset_flow")
	oldPassword := "oldPassword123!"
	newPassword := "newPassword456!"

	// Step 1: Signup
	t.Run("Step 1: Signup", func(t *testing.T) {
		req := generated.RoutesSignupRequestObject{
			Body: &generated.RoutesSignupJSONRequestBody{
				Email:    email,
				Password: oldPassword,
			},
		}

		resp, err := authController.RoutesSignup(ctx, req)
		require.NoError(t, err)
		_, ok := resp.(generated.RoutesSignup201JSONResponse)
		require.True(t, ok)
	})

	// Step 2: Request password reset
	var resetToken string
	t.Run("Step 2: Request password reset", func(t *testing.T) {
		// Use service directly to get the token (in production, this would be sent via email)
		token, err := authService.RequestPasswordReset(ctx, email)
		require.NoError(t, err)
		require.NotEmpty(t, token)
		resetToken = token
	})

	// Step 3: Reset password
	t.Run("Step 3: Reset password", func(t *testing.T) {
		req := generated.RoutesResetPasswordRequestObject{
			Body: &generated.RoutesResetPasswordJSONRequestBody{
				Token:       resetToken,
				NewPassword: newPassword,
			},
		}

		resp, err := authController.RoutesResetPassword(ctx, req)
		require.NoError(t, err)
		_, ok := resp.(generated.RoutesResetPassword200JSONResponse)
		require.True(t, ok)
	})

	// Step 4: Login with new password
	t.Run("Step 4: Login with new password", func(t *testing.T) {
		req := generated.RoutesLoginRequestObject{
			Body: &generated.RoutesLoginJSONRequestBody{
				Email:    email,
				Password: newPassword,
			},
		}

		resp, err := authController.RoutesLogin(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesLogin200JSONResponse)
		require.True(t, ok, "should login with new password")
	})

	// Step 5: Old password no longer works
	t.Run("Step 5: Old password rejected", func(t *testing.T) {
		req := generated.RoutesLoginRequestObject{
			Body: &generated.RoutesLoginJSONRequestBody{
				Email:    email,
				Password: oldPassword,
			},
		}

		resp, err := authController.RoutesLogin(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesLogin401JSONResponse)
		require.True(t, ok, "old password should be rejected")
	})
}

// TestAuthFlow_DuplicateSignup tests that duplicate signups are rejected.
func TestAuthFlow_DuplicateSignup(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	authService := service.NewAuthService(client)
	jwtService := service.NewJWTService("test-secret")
	authController := controller.NewAuthController(authService, jwtService)

	ctx := context.Background()
	email := uniqueEmail("duplicate_signup")
	password := "password123!"

	// First signup succeeds
	t.Run("First signup succeeds", func(t *testing.T) {
		req := generated.RoutesSignupRequestObject{
			Body: &generated.RoutesSignupJSONRequestBody{
				Email:    email,
				Password: password,
			},
		}

		resp, err := authController.RoutesSignup(ctx, req)
		require.NoError(t, err)
		_, ok := resp.(generated.RoutesSignup201JSONResponse)
		require.True(t, ok)
	})

	// Second signup with same email fails
	t.Run("Duplicate signup fails", func(t *testing.T) {
		req := generated.RoutesSignupRequestObject{
			Body: &generated.RoutesSignupJSONRequestBody{
				Email:    email,
				Password: "differentPassword",
			},
		}

		resp, err := authController.RoutesSignup(ctx, req)
		require.NoError(t, err)
		_, ok := resp.(generated.RoutesSignup409JSONResponse)
		require.True(t, ok, "duplicate email should return 409")
	})
}
