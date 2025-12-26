package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/internal/service"
	"github.com/mindhit/api/internal/testutil"
)

// uniqueAuthEmail generates a unique email for auth service tests
func uniqueAuthEmail(prefix string) string {
	return fmt.Sprintf("%s-%s@example.com", prefix, uuid.New().String()[:8])
}

func setupAuthServiceTest(t *testing.T) (*ent.Client, *service.AuthService) {
	client := testutil.SetupTestDB(t)
	authService := service.NewAuthService(client)
	return client, authService
}

func TestAuthService_Signup_Success(t *testing.T) {
	client, authService := setupAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	email := uniqueAuthEmail("signup")

	user, err := authService.Signup(ctx, email, "password123")

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, email, user.Email)
	assert.NotEmpty(t, user.PasswordHash)
	assert.NotEqual(t, "password123", user.PasswordHash) // Password should be hashed
}

func TestAuthService_Signup_DuplicateEmail(t *testing.T) {
	client, authService := setupAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	email := uniqueAuthEmail("duplicate")

	_, err := authService.Signup(ctx, email, "password123")
	require.NoError(t, err)

	_, err = authService.Signup(ctx, email, "password456")

	assert.ErrorIs(t, err, service.ErrEmailExists)
}

func TestAuthService_Login_Success(t *testing.T) {
	client, authService := setupAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	email := uniqueAuthEmail("login")

	_, err := authService.Signup(ctx, email, "password123")
	require.NoError(t, err)

	user, err := authService.Login(ctx, email, "password123")

	require.NoError(t, err)
	assert.Equal(t, email, user.Email)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	client, authService := setupAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	email := uniqueAuthEmail("wrongpwd")

	_, err := authService.Signup(ctx, email, "password123")
	require.NoError(t, err)

	_, err = authService.Login(ctx, email, "wrongpassword")

	assert.ErrorIs(t, err, service.ErrInvalidCredentials)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	client, authService := setupAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()

	_, err := authService.Login(ctx, uniqueAuthEmail("nonexistent"), "password123")

	assert.ErrorIs(t, err, service.ErrInvalidCredentials)
}

func TestAuthService_GetUserByID_Success(t *testing.T) {
	client, authService := setupAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	email := uniqueAuthEmail("getbyid")

	createdUser, err := authService.Signup(ctx, email, "password123")
	require.NoError(t, err)

	user, err := authService.GetUserByID(ctx, createdUser.ID)

	require.NoError(t, err)
	assert.Equal(t, createdUser.ID, user.ID)
	assert.Equal(t, email, user.Email)
}

func TestAuthService_GetUserByID_NotFound(t *testing.T) {
	client, authService := setupAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	nonExistentID := uuid.New()

	_, err := authService.GetUserByID(ctx, nonExistentID)

	assert.ErrorIs(t, err, service.ErrUserNotFound)
}

func TestAuthService_GetUserByEmail_Success(t *testing.T) {
	client, authService := setupAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	email := uniqueAuthEmail("getbyemail")

	_, err := authService.Signup(ctx, email, "password123")
	require.NoError(t, err)

	user, err := authService.GetUserByEmail(ctx, email)

	require.NoError(t, err)
	assert.Equal(t, email, user.Email)
}

func TestAuthService_GetUserByEmail_NotFound(t *testing.T) {
	client, authService := setupAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()

	_, err := authService.GetUserByEmail(ctx, uniqueAuthEmail("notfound"))

	assert.ErrorIs(t, err, service.ErrUserNotFound)
}

// ==================== Password Reset Tests ====================

func TestAuthService_RequestPasswordReset_Success(t *testing.T) {
	client, authService := setupAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	email := uniqueAuthEmail("reset")

	// Create user first
	_, err := authService.Signup(ctx, email, "password123")
	require.NoError(t, err)

	// Request password reset
	token, err := authService.RequestPasswordReset(ctx, email)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Len(t, token, 64) // 32 bytes hex encoded
}

func TestAuthService_RequestPasswordReset_NonExistentEmail(t *testing.T) {
	client, authService := setupAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()

	// Request password reset for non-existent email
	token, err := authService.RequestPasswordReset(ctx, uniqueAuthEmail("nonexistent"))

	// Should return empty string without error (security - prevent email enumeration)
	require.NoError(t, err)
	assert.Empty(t, token)
}

func TestAuthService_RequestPasswordReset_InvalidatesExistingTokens(t *testing.T) {
	client, authService := setupAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	email := uniqueAuthEmail("multi-reset")

	// Create user
	_, err := authService.Signup(ctx, email, "password123")
	require.NoError(t, err)

	// Request first token
	token1, err := authService.RequestPasswordReset(ctx, email)
	require.NoError(t, err)
	assert.NotEmpty(t, token1)

	// Request second token
	token2, err := authService.RequestPasswordReset(ctx, email)
	require.NoError(t, err)
	assert.NotEmpty(t, token2)

	// Tokens should be different
	assert.NotEqual(t, token1, token2)

	// First token should be invalidated (can't be used)
	err = authService.ResetPassword(ctx, token1, "newpassword")
	assert.ErrorIs(t, err, service.ErrTokenInvalid)

	// Second token should work
	err = authService.ResetPassword(ctx, token2, "newpassword")
	require.NoError(t, err)
}

func TestAuthService_ResetPassword_Success(t *testing.T) {
	client, authService := setupAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	email := uniqueAuthEmail("reset-success")

	// Create user
	_, err := authService.Signup(ctx, email, "oldpassword")
	require.NoError(t, err)

	// Request password reset
	token, err := authService.RequestPasswordReset(ctx, email)
	require.NoError(t, err)

	// Reset password
	err = authService.ResetPassword(ctx, token, "newpassword")
	require.NoError(t, err)

	// Verify old password doesn't work
	_, err = authService.Login(ctx, email, "oldpassword")
	assert.ErrorIs(t, err, service.ErrInvalidCredentials)

	// Verify new password works
	user, err := authService.Login(ctx, email, "newpassword")
	require.NoError(t, err)
	assert.Equal(t, email, user.Email)
}

func TestAuthService_ResetPassword_InvalidToken(t *testing.T) {
	client, authService := setupAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()

	err := authService.ResetPassword(ctx, "invalid-token-that-does-not-exist", "newpassword")

	assert.ErrorIs(t, err, service.ErrTokenInvalid)
}

func TestAuthService_ResetPassword_TokenAlreadyUsed(t *testing.T) {
	client, authService := setupAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	email := uniqueAuthEmail("used-token")

	// Create user
	_, err := authService.Signup(ctx, email, "password123")
	require.NoError(t, err)

	// Request password reset
	token, err := authService.RequestPasswordReset(ctx, email)
	require.NoError(t, err)

	// Use token once
	err = authService.ResetPassword(ctx, token, "newpassword1")
	require.NoError(t, err)

	// Try to use same token again
	err = authService.ResetPassword(ctx, token, "newpassword2")

	assert.ErrorIs(t, err, service.ErrTokenInvalid)
}

func TestAuthService_ResetPassword_ExpiredToken(t *testing.T) {
	client, authService := setupAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()
	email := uniqueAuthEmail("expired")

	// Create user
	user, err := authService.Signup(ctx, email, "password123")
	require.NoError(t, err)

	// Create expired token directly in DB (expired 1 hour ago)
	// Use a unique token with UUID to avoid collisions
	expiredTokenStr := fmt.Sprintf("expired-%s", uuid.New().String())
	expiredToken, err := client.PasswordResetToken.Create().
		SetToken(expiredTokenStr).
		SetUserID(user.ID).
		SetExpiresAt(time.Now().Add(-1 * time.Hour)). // Already expired
		Save(ctx)
	require.NoError(t, err)

	// Try to use expired token
	err = authService.ResetPassword(ctx, expiredToken.Token, "newpassword")

	assert.ErrorIs(t, err, service.ErrTokenExpired)
}
