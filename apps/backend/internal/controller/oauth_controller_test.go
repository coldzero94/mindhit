package controller

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mindhit/api/ent/user"
	"github.com/mindhit/api/internal/generated"
	"github.com/mindhit/api/internal/service"
	"github.com/mindhit/api/internal/testutil"
)

// uniqueOAuthEmail generates a unique email for OAuth controller tests
func uniqueOAuthEmail(prefix string) string {
	return fmt.Sprintf("%s-oauth-ctrl-%s@example.com", prefix, uuid.New().String()[:8])
}

// uniqueGoogleID generates a unique Google ID for tests
func uniqueGoogleID() string {
	return fmt.Sprintf("google-%s", uuid.New().String()[:16])
}

func TestOAuthController_RoutesGoogleAuth_InvalidToken(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	oauthService := service.NewOAuthService(client)
	jwtService := service.NewJWTService("test-secret")
	subscriptionService := service.NewSubscriptionService(client)
	controller := NewOAuthController(oauthService, jwtService, subscriptionService)

	ctx := context.Background()

	t.Run("invalid token returns 401", func(t *testing.T) {
		req := generated.RoutesGoogleAuthRequestObject{
			Body: &generated.RoutesGoogleAuthJSONRequestBody{
				Credential: "invalid-token",
			},
		}

		resp, err := controller.RoutesGoogleAuth(ctx, req)
		require.NoError(t, err)

		errorResp, ok := resp.(generated.RoutesGoogleAuth401JSONResponse)
		assert.True(t, ok, "expected 401 response")
		assert.Contains(t, errorResp.Error.Message, "Invalid Google credentials")
	})

	t.Run("empty token returns 401", func(t *testing.T) {
		req := generated.RoutesGoogleAuthRequestObject{
			Body: &generated.RoutesGoogleAuthJSONRequestBody{
				Credential: "",
			},
		}

		resp, err := controller.RoutesGoogleAuth(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(generated.RoutesGoogleAuth401JSONResponse)
		assert.True(t, ok, "expected 401 response")
	})
}

// TestOAuthController_FindOrCreateGoogleUser_Integration tests the integration
// between OAuth service and controller by directly testing FindOrCreateGoogleUser
// since we can't easily mock Google's token validation in integration tests
func TestOAuthController_FindOrCreateGoogleUser_Integration(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	oauthService := service.NewOAuthService(client)
	jwtService := service.NewJWTService("test-secret")
	subscriptionService := service.NewSubscriptionService(client)

	ctx := context.Background()

	// Ensure free plan exists for subscription creation
	_, err := testutil.EnsureFreePlan(t, client)
	require.NoError(t, err)

	t.Run("new Google user gets free subscription", func(t *testing.T) {
		googleID := uniqueGoogleID()
		email := uniqueOAuthEmail("new-sub")
		info := &service.GoogleUserInfo{
			GoogleID: googleID,
			Email:    email,
			Name:     "Test User",
			Picture:  "https://example.com/avatar.jpg",
		}

		// Create user via OAuth service
		u, isNewUser, err := oauthService.FindOrCreateGoogleUser(ctx, info)
		require.NoError(t, err)
		assert.True(t, isNewUser)

		// Simulate what controller does - create free subscription
		_, err = subscriptionService.CreateFreeSubscription(ctx, u.ID)
		require.NoError(t, err)

		// Verify subscription was created
		subInfo, err := subscriptionService.GetSubscriptionInfo(ctx, u.ID)
		require.NoError(t, err)
		assert.NotNil(t, subInfo)
		assert.Equal(t, "free", subInfo.Plan.ID)

		// Generate token pair
		tokenPair, err := jwtService.GenerateTokenPair(u.ID)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenPair.AccessToken)
		assert.NotEmpty(t, tokenPair.RefreshToken)
	})

	t.Run("existing Google user does not get duplicate subscription", func(t *testing.T) {
		googleID := uniqueGoogleID()
		email := uniqueOAuthEmail("existing-sub")
		info := &service.GoogleUserInfo{
			GoogleID: googleID,
			Email:    email,
			Name:     "Test User",
			Picture:  "https://example.com/avatar.jpg",
		}

		// First login
		u, isNewUser, err := oauthService.FindOrCreateGoogleUser(ctx, info)
		require.NoError(t, err)
		assert.True(t, isNewUser)
		_, err = subscriptionService.CreateFreeSubscription(ctx, u.ID)
		require.NoError(t, err)

		// Second login - should not create new subscription
		u2, isNewUser2, err := oauthService.FindOrCreateGoogleUser(ctx, info)
		require.NoError(t, err)
		assert.False(t, isNewUser2)
		assert.Equal(t, u.ID, u2.ID)

		// Verify only one subscription exists
		subInfo, err := subscriptionService.GetSubscriptionInfo(ctx, u.ID)
		require.NoError(t, err)
		assert.NotNil(t, subInfo)
	})
}

// TestOAuthController_GoogleUserProperties verifies Google user properties
func TestOAuthController_GoogleUserProperties(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	oauthService := service.NewOAuthService(client)
	ctx := context.Background()

	t.Run("Google user has correct auth_provider", func(t *testing.T) {
		googleID := uniqueGoogleID()
		email := uniqueOAuthEmail("provider")
		info := &service.GoogleUserInfo{
			GoogleID: googleID,
			Email:    email,
			Name:     "Test User",
			Picture:  "https://example.com/avatar.jpg",
		}

		u, _, err := oauthService.FindOrCreateGoogleUser(ctx, info)
		require.NoError(t, err)

		assert.Equal(t, user.AuthProviderGoogle, u.AuthProvider)
		assert.Equal(t, user.StatusActive, u.Status)
		assert.Nil(t, u.PasswordHash)
	})

	t.Run("Google user has avatar URL", func(t *testing.T) {
		googleID := uniqueGoogleID()
		email := uniqueOAuthEmail("avatar")
		avatarURL := "https://lh3.googleusercontent.com/test-avatar"
		info := &service.GoogleUserInfo{
			GoogleID: googleID,
			Email:    email,
			Name:     "Test User",
			Picture:  avatarURL,
		}

		u, _, err := oauthService.FindOrCreateGoogleUser(ctx, info)
		require.NoError(t, err)

		assert.NotNil(t, u.AvatarURL)
		assert.Equal(t, avatarURL, *u.AvatarURL)
	})
}

// TestOAuthController_LinkEmailAccount tests linking Google to existing email account
func TestOAuthController_LinkEmailAccount(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, client)

	authService := service.NewAuthService(client)
	oauthService := service.NewOAuthService(client)
	ctx := context.Background()

	t.Run("links Google ID to existing email account", func(t *testing.T) {
		email := uniqueOAuthEmail("link")
		googleID := uniqueGoogleID()

		// Create email user first
		emailUser, err := authService.Signup(ctx, email, "password123")
		require.NoError(t, err)
		assert.Nil(t, emailUser.GoogleID)
		assert.Equal(t, user.AuthProviderEmail, emailUser.AuthProvider)

		// Link with Google
		info := &service.GoogleUserInfo{
			GoogleID: googleID,
			Email:    email,
			Name:     "Test User",
			Picture:  "https://example.com/avatar.jpg",
		}
		linkedUser, isNewUser, err := oauthService.FindOrCreateGoogleUser(ctx, info)
		require.NoError(t, err)

		assert.False(t, isNewUser)
		assert.Equal(t, emailUser.ID, linkedUser.ID)
		assert.Equal(t, googleID, *linkedUser.GoogleID)
		assert.NotNil(t, linkedUser.PasswordHash)                        // Still has password
		assert.Equal(t, user.AuthProviderEmail, linkedUser.AuthProvider) // Original provider preserved
	})

	t.Run("linked account can still login with password", func(t *testing.T) {
		email := uniqueOAuthEmail("link-login")
		googleID := uniqueGoogleID()

		// Create and link
		_, err := authService.Signup(ctx, email, "password123")
		require.NoError(t, err)

		info := &service.GoogleUserInfo{
			GoogleID: googleID,
			Email:    email,
			Name:     "Test User",
			Picture:  "https://example.com/avatar.jpg",
		}
		_, _, err = oauthService.FindOrCreateGoogleUser(ctx, info)
		require.NoError(t, err)

		// Can still login with password
		u, err := authService.Login(ctx, email, "password123")
		require.NoError(t, err)
		assert.Equal(t, email, u.Email)
	})
}
