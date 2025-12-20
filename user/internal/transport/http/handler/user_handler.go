// internal/transport/handler/user_handler.go
package handler

import (
	"net/http"
	"strconv"
	"user/internal/app/dto"
	"user/internal/app/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// CreateUser обрабатывает POST /users
// @Summary Создать нового пользователя
// @Description Создание нового пользователя в системе с валидацией email и хешированием пароля
// @Tags users
// @Accept json
// @Produce json
// @Param request body dto.CreateUserRequest true "Данные для создания пользователя"
// @Success 201 {object} dto.UserResponse "Пользователь успешно создан"
// @Failure 400 {object} map[string]string "Некорректные входные данные"
// @Failure 409 {object} map[string]string "Пользователь с таким email уже существует"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/users [post]
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

// GetUser обрабатывает GET /users/:id
// @Summary Получить пользователя по ID
// @Description Получение информации о пользователе по его идентификатору
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 200 {object} dto.UserResponse "Пользователь найден"
// @Failure 400 {object} map[string]string "Некорректный ID"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Router /api/users/{id} [get]
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

// GetUserByEmail обрабатывает GET /users/email/:email
// @Summary Получить пользователя по email
// @Description Поиск пользователя по email адресу
// @Tags users
// @Accept json
// @Produce json
// @Param email path string true "Email пользователя"
// @Success 200 {object} dto.UserResponse "Пользователь найден"
// @Failure 400 {object} map[string]string "Email не указан"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Router /api/users/email/{email} [get]
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

// GetAllUsers обрабатывает GET /users
// @Summary Получить всех пользователей
// @Description Получение списка всех пользователей системы
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} dto.UserResponse "Список пользователей"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/users [get]
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

// UpdateUser обрабатывает PUT /users/:id
// @Summary Обновить пользователя
// @Description Частичное обновление данных пользователя. Можно обновлять одно или несколько полей
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Param request body dto.UpdateUserRequest true "Данные для обновления (можно указать только изменяемые поля)"
// @Success 200 {object} dto.UserResponse "Пользователь успешно обновлён"
// @Failure 400 {object} map[string]string "Некорректные данные или ID"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Failure 409 {object} map[string]string "Email уже используется другим пользователем"
// @Router /api/users/{id} [put]
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

// DeleteUser обрабатывает DELETE /users/:id
// @Summary Удалить пользователя
// @Description Удаление пользователя по ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 200 {object} map[string]interface{} "Пользователь удалён"
// @Failure 400 {object} map[string]string "Некорректный ID"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Router /api/users/{id} [delete]
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

// ValidateUser обрабатывает POST /users/validate
// @Summary Валидировать пользователя
// @Description Проверка учетных данных пользователя (email и пароль)
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body object{email=string,password=string} true "Учетные данные"
// @Success 200 {object} dto.UserResponse "Учетные данные верны"
// @Failure 400 {object} map[string]string "Некорректные данные"
// @Failure 401 {object} map[string]string "Неверные учетные данные"
// @Router /api/users/validate [post]
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
