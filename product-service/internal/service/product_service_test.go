package service

import (
	"context"
	"errors"
	"testing"

	"github.com/microserviceteam0/bff-gateway/product-service/internal/dto"
	"github.com/microserviceteam0/bff-gateway/product-service/internal/model"
)

type mockProductRepository struct {
	findByIDFunc func(ctx context.Context, id int64) (*model.Product, error)
	findAllFunc  func(ctx context.Context) ([]*model.Product, error)
	createFunc   func(ctx context.Context, product *model.Product) error
	updateFunc   func(ctx context.Context, product *model.Product) error
	deleteFunc   func(ctx context.Context, id int64) error
}

func (m *mockProductRepository) FindByID(ctx context.Context, id int64) (*model.Product, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockProductRepository) FindAll(ctx context.Context) ([]*model.Product, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *mockProductRepository) Create(ctx context.Context, product *model.Product) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, product)
	}
	return errors.New("not implemented")
}

func (m *mockProductRepository) Update(ctx context.Context, product *model.Product) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, product)
	}
	return errors.New("not implemented")
}

func (m *mockProductRepository) Delete(ctx context.Context, id int64) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return errors.New("not implemented")
}

func TestProductService_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockRepo := &mockProductRepository{
			findByIDFunc: func(ctx context.Context, id int64) (*model.Product, error) {
				return &model.Product{
					ID:          1,
					Name:        "Test Product",
					Description: "Description",
					Price:       100.0,
					Stock:       10,
				}, nil
			},
		}

		service := NewProductService(mockRepo)
		product, err := service.GetByID(ctx, 1)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if product.ID != 1 {
			t.Errorf("Expected ID=1, got %d", product.ID)
		}
		if product.Name != "Test Product" {
			t.Errorf("Expected Name='Test Product', got '%s'", product.Name)
		}
	})

	t.Run("InvalidID", func(t *testing.T) {
		mockRepo := &mockProductRepository{}
		service := NewProductService(mockRepo)

		_, err := service.GetByID(ctx, 0)
		if err == nil {
			t.Error("Expected error for invalid ID, got nil")
		}

		_, err = service.GetByID(ctx, -1)
		if err == nil {
			t.Error("Expected error for negative ID, got nil")
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		mockRepo := &mockProductRepository{
			findByIDFunc: func(ctx context.Context, id int64) (*model.Product, error) {
				return nil, errors.New("product not found")
			},
		}

		service := NewProductService(mockRepo)
		_, err := service.GetByID(ctx, 999)

		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}

func TestProductService_GetAll(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockRepo := &mockProductRepository{
			findAllFunc: func(ctx context.Context) ([]*model.Product, error) {
				return []*model.Product{
					{ID: 1, Name: "Product 1", Price: 100, Stock: 10},
					{ID: 2, Name: "Product 2", Price: 200, Stock: 20},
				}, nil
			},
		}

		service := NewProductService(mockRepo)
		products, err := service.GetAll(ctx)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(products) != 2 {
			t.Errorf("Expected 2 products, got %d", len(products))
		}
	})

	t.Run("EmptyList", func(t *testing.T) {
		mockRepo := &mockProductRepository{
			findAllFunc: func(ctx context.Context) ([]*model.Product, error) {
				return []*model.Product{}, nil
			},
		}

		service := NewProductService(mockRepo)
		products, err := service.GetAll(ctx)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(products) != 0 {
			t.Errorf("Expected 0 products, got %d", len(products))
		}
	})
}

func TestProductService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockRepo := &mockProductRepository{
			createFunc: func(ctx context.Context, product *model.Product) error {
				product.ID = 1
				return nil
			},
		}

		service := NewProductService(mockRepo)
		req := &dto.CreateProductRequest{
			Name:        "New Product",
			Description: "Description",
			Price:       150.0,
			Stock:       5,
		}

		product, err := service.Create(ctx, req)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if product.ID != 1 {
			t.Errorf("Expected ID=1, got %d", product.ID)
		}
		if product.Name != "New Product" {
			t.Errorf("Expected Name='New Product', got '%s'", product.Name)
		}
	})

	t.Run("RepositoryError", func(t *testing.T) {
		mockRepo := &mockProductRepository{
			createFunc: func(ctx context.Context, product *model.Product) error {
				return errors.New("database error")
			},
		}

		service := NewProductService(mockRepo)
		req := &dto.CreateProductRequest{
			Name:  "New Product",
			Price: 150.0,
			Stock: 5,
		}

		_, err := service.Create(ctx, req)
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}

func TestProductService_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockRepo := &mockProductRepository{
			updateFunc: func(ctx context.Context, product *model.Product) error {
				return nil
			},
		}

		service := NewProductService(mockRepo)
		req := &dto.UpdateProductRequest{
			Name:        "Updated Product",
			Description: "Updated Description",
			Price:       200.0,
			Stock:       15,
		}

		product, err := service.Update(ctx, 1, req)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if product.Name != "Updated Product" {
			t.Errorf("Expected Name='Updated Product', got '%s'", product.Name)
		}
	})

	t.Run("InvalidID", func(t *testing.T) {
		mockRepo := &mockProductRepository{}
		service := NewProductService(mockRepo)
		req := &dto.UpdateProductRequest{
			Name:  "Updated Product",
			Price: 200.0,
			Stock: 15,
		}

		_, err := service.Update(ctx, 0, req)
		if err == nil {
			t.Error("Expected error for invalid ID, got nil")
		}
	})
}

func TestProductService_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockRepo := &mockProductRepository{
			deleteFunc: func(ctx context.Context, id int64) error {
				return nil
			},
		}

		service := NewProductService(mockRepo)
		err := service.Delete(ctx, 1)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("InvalidID", func(t *testing.T) {
		mockRepo := &mockProductRepository{}
		service := NewProductService(mockRepo)

		err := service.Delete(ctx, 0)
		if err == nil {
			t.Error("Expected error for invalid ID, got nil")
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		mockRepo := &mockProductRepository{
			deleteFunc: func(ctx context.Context, id int64) error {
				return errors.New("product not found")
			},
		}

		service := NewProductService(mockRepo)
		err := service.Delete(ctx, 999)

		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}
