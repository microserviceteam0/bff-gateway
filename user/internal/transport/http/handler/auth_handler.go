package handler

import (
	"net/http"
	"user/internal/app/auth_service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService auth_service.AuthService
}

func NewAuthHandler(authService auth_service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) RegisterRoutes(r *gin.Engine) {
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/validate", h.Validate)
	}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginResponse struct {
	Token string `json:"token"`
	Type  string `json:"type"`
}

type RefreshRequest struct {
	Token string `json:"token" binding:"required"`
}

type ValidateRequest struct {
	Token string `json:"token" binding:"required"`
}

type ValidateResponse struct {
	Valid  bool   `json:"valid"`
	UserID int64  `json:"user_id,omitempty"`
	Email  string `json:"email,omitempty"`
	Role   string `json:"role,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
		})
		return
	}

	token, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "AUTHENTICATION_FAILED",
			Message: "Invalid email or password",
		})
		return
	}

	response := LoginResponse{
		Token: token,
		Type:  "Bearer",
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
		})
		return
	}

	newToken, err := h.authService.RefreshToken(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "TOKEN_INVALID",
			Message: "Invalid or expired token",
		})
		return
	}

	response := LoginResponse{
		Token: newToken,
		Type:  "Bearer",
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) Validate(c *gin.Context) {
	var req ValidateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
		})
		return
	}

	claims, err := h.authService.ValidateToken(req.Token)
	response := ValidateResponse{}

	if err != nil {
		response.Valid = false
		c.JSON(http.StatusOK, response)
		return
	}

	response.Valid = true
	response.UserID = claims.UserID
	response.Email = claims.Email
	response.Role = claims.Role

	c.JSON(http.StatusOK, response)
}
