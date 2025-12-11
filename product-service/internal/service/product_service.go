package service

import (
	"context"
	"fmt"

	"github.com/microserviceteam0/bff-gateway/product-service/internal/dto"
	"github.com/microserviceteam0/bff-gateway/product-service/internal/repository"
)

type ProductService interface {
	GetByID(ctx context.Context, id int64) (*dto.ProductResponse, error)
	GetAll(ctx context.Context) ([]*dto.ProductResponse, error)
	Create(ctx context.Context, req *dto.CreateProductRequest) (*dto.ProductResponse, error)
	Update(ctx context.Context, id int64, req *dto.UpdateProductRequest) (*dto.ProductResponse, error)
	Delete(ctx context.Context, id int64) error
}

type productService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) GetByID(ctx context.Context, id int64) (*dto.ProductResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid product id: %d", id)
	}

	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return dto.ToProductResponse(product), nil
}

func (s *productService) GetAll(ctx context.Context) ([]*dto.ProductResponse, error) {
	products, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	return dto.ToProductResponseList(products), nil
}

func (s *productService) Create(ctx context.Context, req *dto.CreateProductRequest) (*dto.ProductResponse, error) {
	product := req.ToProduct()

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, err
	}

	return dto.ToProductResponse(product), nil
}

func (s *productService) Update(ctx context.Context, id int64, req *dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid product id: %d", id)
	}

	product := req.ToProduct()
	product.ID = id

	if err := s.repo.Update(ctx, product); err != nil {
		return nil, err
	}

	return dto.ToProductResponse(product), nil
}

func (s *productService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid product id: %d", id)
	}

	return s.repo.Delete(ctx, id)
}
