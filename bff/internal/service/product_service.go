package service

import (
	"context"

	"github.com/microserviceteam0/bff-gateway/bff/internal/dto"
)

func (s *bffService) ListProducts(ctx context.Context) ([]*dto.ProductResponseDTO, error) {
	products, err := s.productHTTPClient.ListProducts(ctx)
	if err != nil {
		return nil, err
	}

	dtos := make([]*dto.ProductResponseDTO, 0, len(products))
	for _, p := range products {
		dtos = append(dtos, &dto.ProductResponseDTO{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    p.Stock,
		})
	}

	return dtos, nil
}
