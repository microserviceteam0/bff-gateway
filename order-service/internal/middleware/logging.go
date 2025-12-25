package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()
		bodySize := c.Writer.Size()

		if raw != "" {
			path = path + "?" + raw
		}

		attributes := []any{
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.String("client_ip", clientIP),
			slog.Duration("latency", duration),
			slog.Int("body_size", bodySize),
		}

		if errorMessage != "" {
			attributes = append(attributes, slog.String("error", errorMessage))
		}

		if status >= 500 {
			slog.Error("Request failed", attributes...)
		} else if status >= 400 {
			slog.Warn("Request warning", attributes...)
		} else {
			slog.Info("Request processed", attributes...)
		}
	}
}
