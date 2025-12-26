package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetProducts godoc
// @Summary      List all products
// @Description  Get a list of all products
// @Tags         products
// @Accept       json
// @Produce      json
// @Success      200  {array}   dto.ProductResponseDTO
// @Failure      500  {object}  map[string]string
// @Router       /products [get]
func (h *Handler) GetProducts(c *gin.Context) {
	products, err := h.bffService.ListProducts(c.Request.Context())
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, products)
}
