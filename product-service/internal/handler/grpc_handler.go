package handler

import (
	"context"

	pb "github.com/microserviceteam0/bff-gateway/product-service/api/proto"
	"github.com/microserviceteam0/bff-gateway/product-service/internal/dto"
	"github.com/microserviceteam0/bff-gateway/product-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductGRPCHandler struct {
	pb.UnimplementedProductServiceServer
	service service.ProductService
}

func NewProductGRPCHandler(service service.ProductService) *ProductGRPCHandler {
	return &ProductGRPCHandler{service: service}
}

// GetProduct получает один продукт по ID
func (h *ProductGRPCHandler) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.ProductResponse, error) {
	product, err := h.service.GetByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "product not found: %v", err)
	}

	return &pb.ProductResponse{
		Id:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       int32(product.Stock),
		CreatedAt:   product.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   product.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// GetProducts получает несколько продуктов по списку ID
func (h *ProductGRPCHandler) GetProducts(ctx context.Context, req *pb.GetProductsRequest) (*pb.ProductsResponse, error) {
	allProducts, err := h.service.GetAll(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get products: %v", err)
	}

	requestedIDs := make(map[int64]bool)
	for _, id := range req.Ids {
		requestedIDs[id] = true
	}

	var products []*pb.ProductResponse
	for _, p := range allProducts {
		if requestedIDs[p.ID] {
			products = append(products, &pb.ProductResponse{
				Id:          p.ID,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
				Stock:       int32(p.Stock),
				CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			})
		}
	}

	return &pb.ProductsResponse{Products: products}, nil
}

// CheckStock проверяет наличие товара на складе
func (h *ProductGRPCHandler) CheckStock(ctx context.Context, req *pb.CheckStockRequest) (*pb.CheckStockResponse, error) {
	product, err := h.service.GetByID(ctx, req.ProductId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "product not found: %v", err)
	}

	available := int32(product.Stock) >= req.Quantity

	return &pb.CheckStockResponse{
		Available:    available,
		CurrentStock: int32(product.Stock),
	}, nil
}

// UpdateStock обновляет количество товара на складе
func (h *ProductGRPCHandler) UpdateStock(ctx context.Context, req *pb.UpdateStockRequest) (*pb.UpdateStockResponse, error) {
	product, err := h.service.GetByID(ctx, req.ProductId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "product not found: %v", err)
	}

	newStock := int32(product.Stock) + req.QuantityDelta
	if newStock < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "insufficient stock")
	}

	updateReq := &dto.UpdateProductRequest{
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       int(newStock),
	}

	updatedProduct, err := h.service.Update(ctx, req.ProductId, updateReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update stock: %v", err)
	}

	return &pb.UpdateStockResponse{
		NewStock: int32(updatedProduct.Stock),
	}, nil
}
