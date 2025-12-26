package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/microserviceteam0/bff-gateway/bff/internal/clients"
)

const (
	UserIDKey    = "userID"
	UserRoleKey  = "userRole"
	UserEmailKey = "userEmail"
)

func AuthMiddleware(authClient clients.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		var tokenString string

		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString = parts[1]
		} else if len(parts) == 1 {
			tokenString = parts[0]
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}

		resp, err := authClient.ValidateToken(c.Request.Context(), tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "Auth service unavailable: " + err.Error()})
			return
		}

		if !resp.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Set(UserIDKey, resp.UserID)
		c.Set(UserRoleKey, resp.Role)
		c.Set(UserEmailKey, resp.Email)

		c.Next()
	}
}

