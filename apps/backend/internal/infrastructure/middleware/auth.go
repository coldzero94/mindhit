// Package middleware provides HTTP middleware for the API server.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/mindhit/api/internal/service"
)

const (
	// UserIDKey is the context key for storing user ID
	UserIDKey = "userID"
	// UserEmailKey is the context key for storing user email (for test token)
	UserEmailKey = "userEmail"
)

// Auth creates a JWT authentication middleware
func Auth(jwtService *service.JWTService, authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "missing authorization header",
				},
			})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "invalid authorization header format",
				},
			})
			return
		}

		tokenString := parts[1]

		// Check if it's a test token (development only)
		if jwtService.IsTestToken(tokenString) {
			// Get test user from database
			testUser, err := authService.GetUserByEmail(c.Request.Context(), "test@mindhit.dev")
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": gin.H{
						"code":    "UNAUTHORIZED",
						"message": "test user not found",
					},
				})
				return
			}
			c.Set(UserIDKey, testUser.ID)
			c.Set(UserEmailKey, testUser.Email)
			c.Next()
			return
		}

		// Validate access token
		claims, err := jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "invalid or expired token",
				},
			})
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Next()
	}
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return uuid.UUID{}, false
	}
	id, ok := userID.(uuid.UUID)
	return id, ok
}

// MustGetUserID extracts user ID from context or panics
func MustGetUserID(c *gin.Context) uuid.UUID {
	userID, ok := GetUserID(c)
	if !ok {
		panic("user ID not found in context - auth middleware not applied?")
	}
	return userID
}
