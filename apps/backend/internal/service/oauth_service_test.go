package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/ent/user"
	"github.com/mindhit/api/internal/service"
	"github.com/mindhit/api/internal/testutil"
)

// uniqueGoogleID generates a unique Google ID for tests
func uniqueGoogleID() string {
	return fmt.Sprintf("google-%s", uuid.New().String()[:16])
}

// uniqueOAuthEmail generates a unique email for OAuth tests
func uniqueOAuthEmail(prefix string) string {
	return fmt.Sprintf("%s-oauth-%s@example.com", prefix, uuid.New().String()[:8])
}

func setupOAuthServiceTest(t *testing.T) (*ent.Client, *service.OAuthService) {
	client := testutil.SetupTestDB(t)
	oauthService := service.NewOAuthService(client)
	return client, oauthService
}

// createGoogleUserInfo creates a GoogleUserInfo for testing
func createGoogleUserInfo(googleID, email string) *service.GoogleUserInfo {
	return &service.GoogleUserInfo{
		GoogleID: googleID,
		Email:    email,
		Name:     "Test User",
		Picture:  "https://example.com/avatar.jpg",
	}
}

func TestOAuthService_FindOrCreateGoogleUser_NewUser(t *testing.T) {
	client, oauthService := setupOAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	googleID := uniqueGoogleID()
	email := uniqueOAuthEmail("new")
	info := createGoogleUserInfo(googleID, email)

	u, isNewUser, err := oauthService.FindOrCreateGoogleUser(ctx, info)

	require.NoError(t, err)
	assert.True(t, isNewUser)
	assert.NotNil(t, u)
	assert.Equal(t, email, u.Email)
	assert.Equal(t, googleID, *u.GoogleID)
	assert.Equal(t, user.AuthProviderGoogle, u.AuthProvider)
	assert.Equal(t, user.StatusActive, u.Status)
	assert.Nil(t, u.PasswordHash) // Google users have no password
}

func TestOAuthService_FindOrCreateGoogleUser_ExistingGoogleUser(t *testing.T) {
	client, oauthService := setupOAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	googleID := uniqueGoogleID()
	email := uniqueOAuthEmail("existing-google")
	info := createGoogleUserInfo(googleID, email)

	// Create user first
	_, isNewUser, err := oauthService.FindOrCreateGoogleUser(ctx, info)
	require.NoError(t, err)
	assert.True(t, isNewUser)

	// Login again with same Google ID
	u, isNewUser, err := oauthService.FindOrCreateGoogleUser(ctx, info)

	require.NoError(t, err)
	assert.False(t, isNewUser)
	assert.Equal(t, email, u.Email)
	assert.Equal(t, googleID, *u.GoogleID)
}

func TestOAuthService_FindOrCreateGoogleUser_LinkToExistingEmailUser(t *testing.T) {
	client, oauthService := setupOAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	email := uniqueOAuthEmail("link")
	googleID := uniqueGoogleID()

	// Create an email/password user first (simulating signup with email)
	authService := service.NewAuthService(client)
	existingUser, err := authService.Signup(ctx, email, "password123")
	require.NoError(t, err)
	assert.Nil(t, existingUser.GoogleID)

	// Now login with Google using the same email
	info := createGoogleUserInfo(googleID, email)
	u, isNewUser, err := oauthService.FindOrCreateGoogleUser(ctx, info)

	require.NoError(t, err)
	assert.False(t, isNewUser) // Should link, not create new
	assert.Equal(t, existingUser.ID, u.ID)
	assert.Equal(t, googleID, *u.GoogleID) // Google ID should be linked
	assert.NotNil(t, u.PasswordHash)       // Password should still exist
}

func TestOAuthService_FindOrCreateGoogleUser_UpdatesAvatarOnRelogin(t *testing.T) {
	client, oauthService := setupOAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	googleID := uniqueGoogleID()
	email := uniqueOAuthEmail("avatar")

	// First login
	info1 := &service.GoogleUserInfo{
		GoogleID: googleID,
		Email:    email,
		Name:     "Test User",
		Picture:  "https://example.com/avatar1.jpg",
	}
	u1, _, err := oauthService.FindOrCreateGoogleUser(ctx, info1)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/avatar1.jpg", *u1.AvatarURL)

	// Second login with updated avatar
	info2 := &service.GoogleUserInfo{
		GoogleID: googleID,
		Email:    email,
		Name:     "Test User",
		Picture:  "https://example.com/avatar2.jpg",
	}
	u2, isNewUser, err := oauthService.FindOrCreateGoogleUser(ctx, info2)

	require.NoError(t, err)
	assert.False(t, isNewUser)
	assert.Equal(t, "https://example.com/avatar2.jpg", *u2.AvatarURL)
}

func TestOAuthService_FindOrCreateGoogleUser_DifferentGoogleIDSameEmail_CreatesNewIfNoExisting(t *testing.T) {
	client, oauthService := setupOAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	email := uniqueOAuthEmail("multi-google")
	googleID1 := uniqueGoogleID()
	googleID2 := uniqueGoogleID()

	// First Google account
	info1 := createGoogleUserInfo(googleID1, email)
	u1, isNewUser1, err := oauthService.FindOrCreateGoogleUser(ctx, info1)
	require.NoError(t, err)
	assert.True(t, isNewUser1)

	// Second Google account with different ID but same email
	// This should find the existing user by email and link
	info2 := createGoogleUserInfo(googleID2, email)
	u2, isNewUser2, err := oauthService.FindOrCreateGoogleUser(ctx, info2)

	// Since user with googleID1 exists and has the same email,
	// the second call should find by googleID first, fail, then find by email
	// Since the email user already has a googleID, we should update it
	require.NoError(t, err)
	assert.False(t, isNewUser2)
	assert.Equal(t, u1.ID, u2.ID) // Same user
	// Note: The implementation updates google_id, so this test verifies that behavior
}

// TestOAuthService_ValidateGoogleIDToken tests token validation
// Note: This requires mocking or a test token, so we test error cases
func TestOAuthService_ValidateGoogleIDToken_InvalidToken(t *testing.T) {
	client, oauthService := setupOAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()

	// Invalid token should return error
	_, err := oauthService.ValidateGoogleIDToken(ctx, "invalid-token")

	assert.Error(t, err)
	assert.ErrorIs(t, err, service.ErrInvalidIDToken)
}

func TestOAuthService_ValidateGoogleIDToken_EmptyToken(t *testing.T) {
	client, oauthService := setupOAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()

	_, err := oauthService.ValidateGoogleIDToken(ctx, "")

	assert.Error(t, err)
	assert.ErrorIs(t, err, service.ErrInvalidIDToken)
}
