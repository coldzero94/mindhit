package service_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mindhit/api/internal/service"
)

func TestJWTService_GenerateTokenPair(t *testing.T) {
	jwtService := service.NewJWTService("test-secret-key")
	userID := uuid.New()

	tokenPair, err := jwtService.GenerateTokenPair(userID)

	require.NoError(t, err)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.Equal(t, int64(900), tokenPair.ExpiresIn) // 15 minutes = 900 seconds
}

func TestJWTService_ValidateAccessToken(t *testing.T) {
	jwtService := service.NewJWTService("test-secret-key")
	userID := uuid.New()

	tokenPair, err := jwtService.GenerateTokenPair(userID)
	require.NoError(t, err)

	claims, err := jwtService.ValidateAccessToken(tokenPair.AccessToken)

	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, service.AccessToken, claims.TokenType)
}

func TestJWTService_ValidateRefreshToken(t *testing.T) {
	jwtService := service.NewJWTService("test-secret-key")
	userID := uuid.New()

	tokenPair, err := jwtService.GenerateTokenPair(userID)
	require.NoError(t, err)

	claims, err := jwtService.ValidateRefreshToken(tokenPair.RefreshToken)

	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, service.RefreshToken, claims.TokenType)
}

func TestJWTService_ValidateAccessToken_WrongTokenType(t *testing.T) {
	jwtService := service.NewJWTService("test-secret-key")
	userID := uuid.New()

	tokenPair, err := jwtService.GenerateTokenPair(userID)
	require.NoError(t, err)

	// Try to validate refresh token as access token
	_, err = jwtService.ValidateAccessToken(tokenPair.RefreshToken)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected access token")
}

func TestJWTService_ValidateRefreshToken_WrongTokenType(t *testing.T) {
	jwtService := service.NewJWTService("test-secret-key")
	userID := uuid.New()

	tokenPair, err := jwtService.GenerateTokenPair(userID)
	require.NoError(t, err)

	// Try to validate access token as refresh token
	_, err = jwtService.ValidateRefreshToken(tokenPair.AccessToken)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected refresh token")
}

func TestJWTService_InvalidToken(t *testing.T) {
	jwtService := service.NewJWTService("test-secret-key")

	_, err := jwtService.ValidateToken("invalid-token")

	assert.Error(t, err)
}

func TestJWTService_WrongSecret(t *testing.T) {
	jwtService1 := service.NewJWTService("secret-1")
	jwtService2 := service.NewJWTService("secret-2")
	userID := uuid.New()

	tokenPair, err := jwtService1.GenerateTokenPair(userID)
	require.NoError(t, err)

	// Try to validate with different secret
	_, err = jwtService2.ValidateToken(tokenPair.AccessToken)

	assert.Error(t, err)
}

func TestJWTService_GenerateAccessToken(t *testing.T) {
	jwtService := service.NewJWTService("test-secret-key")
	userID := uuid.New()

	token, expiresIn, err := jwtService.GenerateAccessToken(userID)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, int64(900), expiresIn)

	claims, err := jwtService.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
}

func TestJWTService_TokenContainsCorrectClaims(t *testing.T) {
	jwtService := service.NewJWTService("test-secret-key")
	userID := uuid.New()

	tokenPair, err := jwtService.GenerateTokenPair(userID)
	require.NoError(t, err)

	claims, err := jwtService.ValidateToken(tokenPair.AccessToken)
	require.NoError(t, err)

	// Check registered claims
	assert.Equal(t, "mindhit", claims.Issuer)
	assert.Equal(t, userID.String(), claims.Subject)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
	assert.NotEmpty(t, claims.ID) // jti should be set

	// Check expiration is approximately 15 minutes from now
	expectedExpiry := time.Now().Add(15 * time.Minute)
	assert.WithinDuration(t, expectedExpiry, claims.ExpiresAt.Time, 5*time.Second)
}

func TestJWTService_TokensAreUniqueEvenInSameSecond(t *testing.T) {
	jwtService := service.NewJWTService("test-secret-key")
	userID := uuid.New()

	// Generate multiple tokens for the same user rapidly
	token1, _, err := jwtService.GenerateAccessToken(userID)
	require.NoError(t, err)

	token2, _, err := jwtService.GenerateAccessToken(userID)
	require.NoError(t, err)

	token3, _, err := jwtService.GenerateAccessToken(userID)
	require.NoError(t, err)

	// All tokens should be different due to unique jti
	assert.NotEqual(t, token1, token2, "tokens generated in same second should be unique")
	assert.NotEqual(t, token2, token3, "tokens generated in same second should be unique")
	assert.NotEqual(t, token1, token3, "tokens generated in same second should be unique")

	// Verify each token has a unique jti
	claims1, err := jwtService.ValidateAccessToken(token1)
	require.NoError(t, err)

	claims2, err := jwtService.ValidateAccessToken(token2)
	require.NoError(t, err)

	claims3, err := jwtService.ValidateAccessToken(token3)
	require.NoError(t, err)

	assert.NotEqual(t, claims1.ID, claims2.ID, "jti should be unique")
	assert.NotEqual(t, claims2.ID, claims3.ID, "jti should be unique")
	assert.NotEqual(t, claims1.ID, claims3.ID, "jti should be unique")
}
