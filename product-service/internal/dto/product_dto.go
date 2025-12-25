package dto

import (
	"time"

	"github.com/microserviceteam0/bff-gateway/product-service/internal/model"
)

// CreateProductRequest - DTO для создания продукта
type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required,max=255"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gte=0"`
	Stock       int     `json:"stock" validate:"required,gte=0"`
}

// UpdateProductRequest - DTO для обновления продукта
type UpdateProductRequest struct {
	Name        string  `json:"name" validate:"required,max=255"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gte=0"`
	Stock       int     `json:"stock" validate:"required,gte=0"`
}

// ProductResponse - DTO для ответа
type ProductResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToProductResponse конвертирует domain model в DTO
func ToProductResponse(product *model.Product) *ProductResponse {
	return &ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
}

// ToProductResponseList конвертирует список domain models в DTOs
func ToProductResponseList(products []*model.Product) []*ProductResponse {
	responses := make([]*ProductResponse, 0, len(products))
	for _, p := range products {
		responses = append(responses, ToProductResponse(p))
	}
	return responses
}

// ToProduct конвертирует CreateProductRequest в domain model
func (r *CreateProductRequest) ToProduct() *model.Product {
	return &model.Product{
		Name:        r.Name,
		Description: r.Description,
		Price:       r.Price,
		Stock:       r.Stock,
	}
}

// ToProduct конвертирует UpdateProductRequest в domain model
func (r *UpdateProductRequest) ToProduct() *model.Product {
	return &model.Product{
		Name:        r.Name,
		Description: r.Description,
		Price:       r.Price,
		Stock:       r.Stock,
	}
}
