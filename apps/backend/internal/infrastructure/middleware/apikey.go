// Package middleware provides HTTP middleware for the API server.
package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// APIKeyAuth creates an API key authentication middleware.
// It validates requests against the API_SECRET_KEY environment variable.
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := os.Getenv("API_SECRET_KEY")
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "CONFIGURATION_ERROR",
					"message": "API_SECRET_KEY not configured",
				},
			})
			return
		}

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

		if parts[1] != apiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "invalid API key",
				},
			})
			return
		}

		c.Next()
	}
}
