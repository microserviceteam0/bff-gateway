package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func RateLimit(limit float64, burst int) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(limit), burst)
	
	return func(c *gin.Context) {
		if !limiter.Allow() {
			slog.Warn("Rate limit exceeded", "ip", c.ClientIP())
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests",
			})
			return
		}
		c.Next()
	}
}
