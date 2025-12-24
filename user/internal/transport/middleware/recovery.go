package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RecoveryMiddleware перехватывает паники и логирует их
func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := c.Get("request_id")

				// Логируем панику
				logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("request_id", requestID.(string)),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.Stack("stack"),
				)

				// Отправляем клиенту 500 ошибку
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":      "internal_server_error",
					"message":    "Something went wrong",
					"request_id": requestID,
				})
			}
		}()

		c.Next()
	}
}
