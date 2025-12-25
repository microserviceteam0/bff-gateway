package handler

import (
	"net/http"
	"strconv"
	"user/internal/app/dto"
	"user/internal/app/user_service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService user_service.UserService
}

func NewUserHandler(userService user_service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"message": err.Error(),
		})
		return
	}

	user, err := h.userService.CreateUser(
		req.Name,
		req.Email,
		req.Password,
		req.Role,
	)

	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "user with this email already exists" {
			status = http.StatusConflict
		}

		c.JSON(status, gin.H{
			"error":   "creation_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.ToResponse(user))
}

func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "User ID must be a number",
		})
		return
	}

	user, err := h.userService.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "user_not_found",
			"message": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ToResponse(user))
}

func (h *UserHandler) GetUserByEmail(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_email",
			"message": "Email parameter is required",
		})
		return
	}

	user, err := h.userService.GetUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "user_not_found",
			"message": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ToResponse(user))
}

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to fetch users",
		})
		return
	}

	var response []dto.UserResponse
	for _, user := range users {
		response = append(response, dto.ToResponse(&user))
	}

	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "User ID must be a number",
		})
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"message": err.Error(),
		})
		return
	}

	if req.Name == nil && req.Email == nil && req.Role == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "no_fields_to_update",
			"message": "At least one field (name, email, or role) must be provided",
		})
		return
	}

	var name, email, role string
	if req.Name != nil {
		name = *req.Name
	}
	if req.Email != nil {
		email = *req.Email
	}
	if req.Role != nil {
		role = *req.Role
	}

	err = h.userService.UpdateUser(id, name, email, role)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		} else if err.Error() == "email already in use by another user" {
			status = http.StatusConflict
		}

		c.JSON(status, gin.H{
			"error":   "update_failed",
			"message": err.Error(),
		})
		return
	}

	user, err := h.userService.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "User updated successfully",
			"id":      id,
		})
		return
	}

	c.JSON(http.StatusOK, dto.ToResponse(user))
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "User ID must be a number",
		})
		return
	}

	err = h.userService.DeleteUser(id)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		}

		c.JSON(status, gin.H{
			"error":   "delete_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
		"id":      id,
	})
}

func (h *UserHandler) ValidateUser(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"message": err.Error(),
		})
		return
	}

	user, err := h.userService.ValidateCredentials(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "authentication_failed",
			"message": "Invalid email or password",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ToResponse(user))
}
