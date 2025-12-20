package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/microserviceteam0/bff-gateway/product-service/internal/model"
	"github.com/microserviceteam0/bff-gateway/shared/metrics"
)

type ProductRepository interface {
	FindByID(ctx context.Context, id int64) (*model.Product, error)
	FindAll(ctx context.Context) ([]*model.Product, error)
	Create(ctx context.Context, product *model.Product) error
	Update(ctx context.Context, product *model.Product) error
	Delete(ctx context.Context, id int64) error
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) ProductRepository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) FindByID(ctx context.Context, id int64) (*model.Product, error) {
	start := time.Now()

	query := `SELECT id, name, description, price, stock, created_at, updated_at FROM products WHERE id = $1`

	var product model.Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Stock,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	duration := time.Since(start).Seconds()
	metrics.DBQueryDuration.WithLabelValues("product-service", "SELECT").Observe(duration)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("product with id %d not found", id)
	}

	if err != nil {
		metrics.DBErrors.WithLabelValues("product-service", "SELECT").Inc()
		return nil, err
	}

	return &product, nil
}

func (r *postgresRepository) FindAll(ctx context.Context) ([]*model.Product, error) {
	start := time.Now()

	query := `SELECT id, name, description, price, stock, created_at, updated_at FROM products ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		duration := time.Since(start).Seconds()
		metrics.DBQueryDuration.WithLabelValues("product-service", "SELECT").Observe(duration)
		metrics.DBErrors.WithLabelValues("product-service", "SELECT").Inc()
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var products []*model.Product
	for rows.Next() {
		var product model.Product
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Stock,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			metrics.DBErrors.WithLabelValues("product-service", "SELECT").Inc()
			return nil, err
		}
		products = append(products, &product)
	}

	if err = rows.Err(); err != nil {
		metrics.DBErrors.WithLabelValues("product-service", "SELECT").Inc()
		return nil, err
	}

	duration := time.Since(start).Seconds()
	metrics.DBQueryDuration.WithLabelValues("product-service", "SELECT").Observe(duration)

	return products, nil
}

func (r *postgresRepository) Create(ctx context.Context, product *model.Product) error {
	start := time.Now()

	query := `INSERT INTO products (name, description, price, stock) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)

	duration := time.Since(start).Seconds()
	metrics.DBQueryDuration.WithLabelValues("product-service", "INSERT").Observe(duration)

	if err != nil {
		metrics.DBErrors.WithLabelValues("product-service", "INSERT").Inc()
		return err
	}

	return nil
}

func (r *postgresRepository) Update(ctx context.Context, product *model.Product) error {
	start := time.Now()

	query := `UPDATE products SET name = $1, description = $2, price = $3, stock = $4 WHERE id = $5 RETURNING updated_at`

	err := r.db.QueryRowContext(ctx, query,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.ID,
	).Scan(&product.UpdatedAt)

	duration := time.Since(start).Seconds()
	metrics.DBQueryDuration.WithLabelValues("product-service", "UPDATE").Observe(duration)

	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("product with id %d not found", product.ID)
	}

	if err != nil {
		metrics.DBErrors.WithLabelValues("product-service", "UPDATE").Inc()
		return err
	}

	return nil
}

func (r *postgresRepository) Delete(ctx context.Context, id int64) error {
	start := time.Now()

	query := `DELETE FROM products WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)

	duration := time.Since(start).Seconds()
	metrics.DBQueryDuration.WithLabelValues("product-service", "DELETE").Observe(duration)

	if err != nil {
		metrics.DBErrors.WithLabelValues("product-service", "DELETE").Inc()
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		metrics.DBErrors.WithLabelValues("product-service", "DELETE").Inc()
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product with id %d not found", id)
	}

	return nil
}
