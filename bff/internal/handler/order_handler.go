package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/microserviceteam0/bff-gateway/bff/internal/dto"
)

// CreateOrder godoc
// @Summary      Create a new order
// @Description  Create a new order for the authenticated user
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input body dto.CreateOrderRequestDTO true "Order info"
// @Success      201  {object}  dto.OrderResponseDTO
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /orders [post]
func (h *Handler) CreateOrder(c *gin.Context) {
	userID := getUserIDFromContext(c)
	userRole := getUserRoleFromContext(c)

	var req dto.CreateOrderRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.bffService.CreateOrder(c.Request.Context(), userID, userRole, req)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetOrder godoc
// @Summary      Get order details
// @Description  Get order details by ID
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Order ID"
// @Success      200  {object}  dto.OrderResponseDTO
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /orders/{id} [get]
func (h *Handler) GetOrder(c *gin.Context) {
	userID := getUserIDFromContext(c)
	userRole := getUserRoleFromContext(c)
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	resp, err := h.bffService.GetOrderDetails(c.Request.Context(), userID, userRole, id)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CancelOrder godoc
// @Summary      Cancel an order
// @Description  Cancel an order by ID
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Order ID"
// @Param        input body dto.CancelOrderRequestDTO true "Cancellation reason"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /orders/{id}/cancel [post]
func (h *Handler) CancelOrder(c *gin.Context) {
	userID := getUserIDFromContext(c)
	userRole := getUserRoleFromContext(c)
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req dto.CancelOrderRequestDTO
	_ = c.ShouldBindJSON(&req)

	err = h.bffService.CancelOrder(c.Request.Context(), userID, userRole, id, req.Reason)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "cancelled"})
}
