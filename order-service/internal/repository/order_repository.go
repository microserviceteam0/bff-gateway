package repository

import (
	"context"
	"time"

	"order-service/internal/model"

	"github.com/microserviceteam0/bff-gateway/shared/metrics"
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
	start := time.Now()
	err := o.db.
		WithContext(ctx).
		Create(order).
		Error

	duration := time.Since(start).Seconds()
	metrics.DBQueryDuration.WithLabelValues("order-service", "INSERT").Observe(duration)

	if err != nil {
		metrics.DBErrors.WithLabelValues("order-service", "INSERT").Inc()
		return nil, err
	}

	return order, nil
}

// GetOrder implements OrderRepository.
func (o *OrderRepositoryImpl) GetOrder(ctx context.Context, orderID int64) (*model.Order, error) {
	start := time.Now()
	var order model.Order
	err := o.db.
		WithContext(ctx).
		Preload("Items").
		First(&order, orderID).
		Error

	duration := time.Since(start).Seconds()
	metrics.DBQueryDuration.WithLabelValues("order-service", "SELECT").Observe(duration)

	if err != nil {
		metrics.DBErrors.WithLabelValues("order-service", "SELECT").Inc()
		return nil, err
	}
	return &order, err
}

// GetOrdersByUserID implements OrderRepository.
func (o *OrderRepositoryImpl) GetOrdersByUserID(ctx context.Context, userID int64, limit int, offset int) ([]model.Order, int64, error) {
	start := time.Now()
	var orders []model.Order
	var total int64

	query := o.db.WithContext(ctx).Model(&model.Order{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		metrics.DBErrors.WithLabelValues("order-service", "SELECT").Inc()
		return nil, 0, err
	}

	err := query.
		Preload("Items").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).
		Error

	duration := time.Since(start).Seconds()
	metrics.DBQueryDuration.WithLabelValues("order-service", "SELECT").Observe(duration)

	if err != nil {
		metrics.DBErrors.WithLabelValues("order-service", "SELECT").Inc()
		return nil, 0, err
	}
	return orders, total, nil
}

// UpdateOrder implements OrderRepository.
func (o *OrderRepositoryImpl) UpdateOrder(ctx context.Context, order *model.Order) error {
	start := time.Now()
	err := o.db.
		WithContext(ctx).
		Save(order).
		Error

	duration := time.Since(start).Seconds()
	metrics.DBQueryDuration.WithLabelValues("order-service", "UPDATE").Observe(duration)

	if err != nil {
		metrics.DBErrors.WithLabelValues("order-service", "UPDATE").Inc()
		return err
	}
	return nil
}

// Delete implements OrderRepository.
func (o *OrderRepositoryImpl) Delete(ctx context.Context, orderID int64) error {
	start := time.Now()
	err := o.db.
		WithContext(ctx).
		Delete(&model.Order{}, orderID).
		Error

	duration := time.Since(start).Seconds()
	metrics.DBQueryDuration.WithLabelValues("order-service", "DELETE").Observe(duration)

	if err != nil {
		metrics.DBErrors.WithLabelValues("order-service", "DELETE").Inc()
		return err
	}

	return nil
}
