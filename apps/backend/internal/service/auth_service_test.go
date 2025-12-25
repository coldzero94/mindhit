package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mindhit/api/ent"
	"github.com/mindhit/api/internal/service"
	"github.com/mindhit/api/internal/testutil"
)

func setupAuthServiceTest(t *testing.T) (*ent.Client, *service.AuthService) {
	client := testutil.SetupTestDB(t)
	authService := service.NewAuthService(client)
	return client, authService
}

func TestAuthService_Signup_Success(t *testing.T) {
	client, authService := setupAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()

	user, err := authService.Signup(ctx, "test@example.com", "password123")

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Email)
	assert.NotEmpty(t, user.PasswordHash)
	assert.NotEqual(t, "password123", user.PasswordHash) // Password should be hashed
}

func TestAuthService_Signup_DuplicateEmail(t *testing.T) {
	client, authService := setupAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()

	_, err := authService.Signup(ctx, "test@example.com", "password123")
	require.NoError(t, err)

	_, err = authService.Signup(ctx, "test@example.com", "password456")

	assert.ErrorIs(t, err, service.ErrEmailExists)
}

func TestAuthService_Login_Success(t *testing.T) {
	client, authService := setupAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()

	_, err := authService.Signup(ctx, "test@example.com", "password123")
	require.NoError(t, err)

	user, err := authService.Login(ctx, "test@example.com", "password123")

	require.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	client, authService := setupAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()

	_, err := authService.Signup(ctx, "test@example.com", "password123")
	require.NoError(t, err)

	_, err = authService.Login(ctx, "test@example.com", "wrongpassword")

	assert.ErrorIs(t, err, service.ErrInvalidCredentials)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	client, authService := setupAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()

	_, err := authService.Login(ctx, "nonexistent@example.com", "password123")

	assert.ErrorIs(t, err, service.ErrInvalidCredentials)
}

func TestAuthService_GetUserByID_Success(t *testing.T) {
	client, authService := setupAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()

	createdUser, err := authService.Signup(ctx, "test@example.com", "password123")
	require.NoError(t, err)

	user, err := authService.GetUserByID(ctx, createdUser.ID)

	require.NoError(t, err)
	assert.Equal(t, createdUser.ID, user.ID)
	assert.Equal(t, "test@example.com", user.Email)
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

	_, err := authService.Signup(ctx, "test@example.com", "password123")
	require.NoError(t, err)

	user, err := authService.GetUserByEmail(ctx, "test@example.com")

	require.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email)
}

func TestAuthService_GetUserByEmail_NotFound(t *testing.T) {
	client, authService := setupAuthServiceTest(t)
	defer testutil.CleanupTestDB(t, client)

	ctx := context.Background()

	_, err := authService.GetUserByEmail(ctx, "nonexistent@example.com")

	assert.ErrorIs(t, err, service.ErrUserNotFound)
}
