package middleware

import (
	"strings"
	"user/internal/app/auth_service"

	"github.com/gin-gonic/gin"
)

const (
	UserIDKey = "user_id"
	EmailKey  = "email"
	RoleKey   = "role"
)

func AuthMiddleware(authService auth_service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {

		if isPublicRoute(c.Request.URL.Path) {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(401, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(401, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(EmailKey, claims.Email)
		c.Set(RoleKey, claims.Role)

		c.Next()
	}
}

func isPublicRoute(path string) bool {
	publicRoutes := []string{
		"/api/auth/login",
		"/api/auth/refresh",
		"/api/auth/validate",
		"/health",
		"/metrics",
		"/swagger",
	}

	for _, route := range publicRoutes {
		if strings.HasPrefix(path, route) {
			return true
		}
	}
	return false
}

func GetUserID(c *gin.Context) (int64, bool) {
	value, exists := c.Get(UserIDKey)
	if !exists {
		return 0, false
	}
	return value.(int64), true
}

func GetEmail(c *gin.Context) (string, bool) {
	value, exists := c.Get(EmailKey)
	if !exists {
		return "", false
	}
	return value.(string), true
}

func GetRole(c *gin.Context) (string, bool) {
	value, exists := c.Get(RoleKey)
	if !exists {
		return "", false
	}
	return value.(string), true
}
