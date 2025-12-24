package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// HealthCheck проверяет работоспособность сервиса
// @Summary Проверка здоровья сервиса
// @Description Проверка статуса сервиса и подключения к базе данных
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Сервис работает"
// @Success 503 {object} map[string]interface{} "Сервис недоступен"
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	// Проверяем подключение к БД
	dbStatus := "healthy"
	if h.db != nil {
		sqlDB, err := h.db.DB()
		if err != nil {
			dbStatus = "unhealthy"
		} else if err := sqlDB.Ping(); err != nil {
			dbStatus = "unhealthy"
		}
	}

	status := http.StatusOK
	if dbStatus != "healthy" {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"status":       dbStatus,
		"auth_service": "user-auth_service",
		"timestamp":    time.Now().Format(time.RFC3339),
		"version":      "1.0.0",
		"database":     dbStatus,
	})
}
