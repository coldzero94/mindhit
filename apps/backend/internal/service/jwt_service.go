package service

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type JWTService struct {
	secret            []byte
	accessExpiration  time.Duration
	refreshExpiration time.Duration
	isDev             bool
}

type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // Access token expiry in seconds
}

func NewJWTService(secret string) *JWTService {
	isDev := os.Getenv("ENVIRONMENT") != "production"
	return &JWTService{
		secret:            []byte(secret),
		accessExpiration:  15 * time.Minute,
		refreshExpiration: 7 * 24 * time.Hour,
		isDev:             isDev,
	}
}

// GenerateTokenPair creates both access and refresh tokens
func (s *JWTService) GenerateTokenPair(userID uuid.UUID) (*TokenPair, error) {
	accessToken, err := s.generateToken(userID, AccessToken, s.accessExpiration)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.generateToken(userID, RefreshToken, s.refreshExpiration)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessExpiration.Seconds()),
	}, nil
}

// GenerateAccessToken creates only access token (for refresh)
func (s *JWTService) GenerateAccessToken(userID uuid.UUID) (string, int64, error) {
	token, err := s.generateToken(userID, AccessToken, s.accessExpiration)
	if err != nil {
		return "", 0, err
	}
	return token, int64(s.accessExpiration.Seconds()), nil
}

func (s *JWTService) generateToken(userID uuid.UUID, tokenType TokenType, expiration time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(), // jti: unique token identifier
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "mindhit",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ValidateToken validates any token type
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// ValidateRefreshToken validates specifically refresh token
func (s *JWTService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != RefreshToken {
		return nil, fmt.Errorf("invalid token type: expected refresh token")
	}

	return claims, nil
}

// ValidateAccessToken validates specifically access token
func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	// Development environment: allow test token
	if s.isDev {
		testToken := os.Getenv("TEST_ACCESS_TOKEN")
		if testToken != "" && tokenString == testToken {
			// Return claims for test user (will be resolved later in middleware)
			return &Claims{
				TokenType: AccessToken,
			}, nil
		}
	}

	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != AccessToken {
		return nil, fmt.Errorf("invalid token type: expected access token")
	}

	return claims, nil
}

// IsTestToken checks if the token is a test token (for middleware use)
func (s *JWTService) IsTestToken(tokenString string) bool {
	if !s.isDev {
		return false
	}
	testToken := os.Getenv("TEST_ACCESS_TOKEN")
	return testToken != "" && tokenString == testToken
}
