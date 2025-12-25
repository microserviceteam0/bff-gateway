package repository

import (
	"context"

	"order-service/internal/model"

	"gorm.io/gorm"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *model.Order) (*model.Order, error)
	GetOrder(ctx context.Context, orderID int64) (*model.Order, error)
	GetOrdersByUserID(ctx context.Context, userID int64, limit int, offset int) ([]model.Order, int64, error)
	UpdateOrder(ctx context.Context, order *model.Order) error
	Delete(ctx context.Context, orderID int64) error
}

type OrderRepositoryImpl struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &OrderRepositoryImpl{db: db}
}

// CreateOrder implements OrderRepository.
func (o *OrderRepositoryImpl) CreateOrder(ctx context.Context, order *model.Order) (*model.Order, error) {
	err := o.db.
		WithContext(ctx).
		Create(order).
		Error
	if err != nil {
		return nil, err
	}

	return order, nil
}

// GetOrder implements OrderRepository.
func (o *OrderRepositoryImpl) GetOrder(ctx context.Context, orderID int64) (*model.Order, error) {
	var order model.Order
	err := o.db.
		WithContext(ctx).
		Preload("Items").
		First(&order, orderID).
		Error
	if err != nil {
		return nil, err
	}
	return &order, err
}

// GetOrdersByUserID implements OrderRepository.
func (o *OrderRepositoryImpl) GetOrdersByUserID(ctx context.Context, userID int64, limit int, offset int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	query := o.db.WithContext(ctx).Model(&model.Order{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Items").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).
		Error
	if err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

// UpdateOrder implements OrderRepository.
func (o *OrderRepositoryImpl) UpdateOrder(ctx context.Context, order *model.Order) error {
	err := o.db.
		WithContext(ctx).
		Save(order).
		Error
	if err != nil {
		return err
	}
	return nil
}

// Delete implements OrderRepository.
func (o *OrderRepositoryImpl) Delete(ctx context.Context, orderID int64) error {
	err := o.db.
		WithContext(ctx).
		Delete(&model.Order{}, orderID).
		Error
	if err != nil {
		return err
	}

	return nil
}
