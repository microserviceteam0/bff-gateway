package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func RedisCacheMiddleware(rdb *redis.Client, expiration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Кэшируем только GET запросы
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// Создаем уникальный ключ на основе URL, параметров и ID пользователя (если есть)
		userID, _ := c.Get(UserIDKey)
		cacheKey := generateCacheKey(c.Request.RequestURI, userID)

		// Пытаемся получить данные из Redis
		val, err := rdb.Get(c.Request.Context(), cacheKey).Result()
		if err == nil {
			// Если данные есть, возвращаем их
			c.Header("X-Cache", "HIT")
			c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(val))
			c.Abort()
			return
		}

		// Если данных нет, перехватываем ответ
		w := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = w

		c.Next()

		// Если запрос успешен, сохраняем в Redis
		if c.Writer.Status() == http.StatusOK {
			rdb.Set(context.Background(), cacheKey, w.body.Bytes(), expiration)
		}
	}
}

func generateCacheKey(uri string, userID interface{}) string {
	key := uri
	if userID != nil {
		key = fmt.Sprintf("%v:%s", userID, uri)
	}
	hash := sha256.Sum256([]byte(key))
	return "cache:" + hex.EncodeToString(hash[:])
}
