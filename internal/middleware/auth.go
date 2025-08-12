package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func BearerAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}
		// Split the header to get the token part
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			return
		}

		token := parts[1]
		// Validate the token (add your validation logic here)
		if !isValidToken(token) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Next()
	}

}

func isValidToken(token string) bool {
	// Example: validate against a hardcoded token (for demo only)
	// In production, validate against your auth service or JWT
	return token == "gzzg"
}
