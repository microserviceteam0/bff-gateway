package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/microserviceteam0/bff-gateway/bff/internal/dto"
)

// Register godoc
// @Summary      Register a new user
// @Description  Register a new user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body dto.RegisterUserRequestDTO true "User registration info"
// @Success      201  {object}  dto.UserResponseDTO
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /register [post]
func (h *Handler) Register(c *gin.Context) {
	var req dto.RegisterUserRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.bffService.Register(c.Request.Context(), req)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Login godoc
// @Summary      Login user
// @Description  Login user and get token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body dto.LoginRequestDTO true "Login info"
// @Success      200  {object}  dto.LoginResponseDTO
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /login [post]
func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.bffService.Login(c.Request.Context(), req)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetProfile godoc
// @Summary      Get user profile
// @Description  Get current user profile
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.UserResponseDTO
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /profile [get]
func (h *Handler) GetProfile(c *gin.Context) {
	userID := getUserIDFromContext(c)
	userRole := getUserRoleFromContext(c)

	resp, err := h.bffService.GetUserProfile(c.Request.Context(), userID, userRole)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
