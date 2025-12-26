package dto

import "time"

type OrderResponseDTO struct {
	ID        int64          `json:"id"`
	User      UserSummaryDTO `json:"user"`
	Items     []OrderItemDTO `json:"items"`
	Status    string         `json:"status"`
	TotalSum  float64        `json:"total_sum"`
	CreatedAt time.Time      `json:"created_at"`
}

type UserSummaryDTO struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type OrderItemDTO struct {
	ProductID   int64   `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int32   `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

// CreateOrderRequestDTO — запрос на создание заказа
type CreateOrderRequestDTO struct {
	Items []CreateOrderItemDTO `json:"items"`
}

type CreateOrderItemDTO struct {
	ProductID int64 `json:"product_id"`
	Quantity  int32 `json:"quantity"`
}

type CancelOrderRequestDTO struct {
	Reason string `json:"reason"`
}
