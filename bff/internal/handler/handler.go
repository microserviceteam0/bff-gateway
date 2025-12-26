package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/microserviceteam0/bff-gateway/bff/internal/apperr"
	"github.com/microserviceteam0/bff-gateway/bff/internal/middleware"
	"github.com/microserviceteam0/bff-gateway/bff/internal/service"
)

type Handler struct {
	bffService service.BFFService
}

func NewHandler(bffService service.BFFService) *Handler {
	return &Handler{
		bffService: bffService,
	}
}

func getUserIDFromContext(c *gin.Context) int64 {
	id, exists := c.Get(middleware.UserIDKey)
	if !exists {
		return 0
	}
	if val, ok := id.(int64); ok {
		return val
	}
	return 0
}

func getUserRoleFromContext(c *gin.Context) string {
	role, exists := c.Get(middleware.UserRoleKey)
	if !exists {
		return ""
	}
	if val, ok := role.(string); ok {
		return val
	}
	return ""
}

func (h *Handler) respondWithError(c *gin.Context, err error) {
	c.Error(err)

	httpCode := http.StatusInternalServerError
	message := "Internal Server Error"

	if errors.Is(err, apperr.ErrNotFound) {
		httpCode = http.StatusNotFound
		message = err.Error()
	} else if errors.Is(err, apperr.ErrInvalidInput) {
		httpCode = http.StatusBadRequest
		message = err.Error()
	} else if errors.Is(err, apperr.ErrUnauthorized) {
		httpCode = http.StatusUnauthorized
		message = "Unauthorized"
	} else if errors.Is(err, apperr.ErrForbidden) {
		httpCode = http.StatusForbidden
		message = "Access denied"
	} else if errors.Is(err, apperr.ErrAlreadyExists) {
		httpCode = http.StatusConflict
		message = err.Error()
	} else if errors.Is(err, apperr.ErrServiceUnavailable) {
		httpCode = http.StatusServiceUnavailable
		message = "Service unavailable"
	} else if errors.Is(err, apperr.ErrTimeout) {
		httpCode = http.StatusGatewayTimeout
		message = "Request timeout"
	}

	c.JSON(httpCode, gin.H{"error": message})
}

