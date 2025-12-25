package handler

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/microserviceteam0/bff-gateway/product-service/api/proto"
	"github.com/microserviceteam0/bff-gateway/product-service/internal/dto"
	"github.com/microserviceteam0/bff-gateway/product-service/internal/service"
	"github.com/microserviceteam0/bff-gateway/product-service/pkg/logger"
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
	logger.Debug("gRPC GetProduct called",
		zap.Int64("product_id", req.Id),
	)

	product, err := h.service.GetByID(ctx, req.Id)
	if err != nil {
		logger.Error("gRPC GetProduct failed",
			zap.Int64("product_id", req.Id),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.NotFound, "product not found: %v", err)
	}

	logger.Info("gRPC GetProduct success",
		zap.Int64("product_id", product.ID),
		zap.String("product_name", product.Name),
	)

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
	logger.Debug("gRPC GetProducts called",
		zap.Int("ids_count", len(req.Ids)),
		zap.Int64s("product_ids", req.Ids),
	)

	allProducts, err := h.service.GetAll(ctx)
	if err != nil {
		logger.Error("gRPC GetProducts failed",
			zap.Error(err),
		)
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

	logger.Info("gRPC GetProducts success",
		zap.Int("requested_count", len(req.Ids)),
		zap.Int("returned_count", len(products)),
	)

	return &pb.ProductsResponse{Products: products}, nil
}

// CheckStock проверяет наличие товара на складе
func (h *ProductGRPCHandler) CheckStock(ctx context.Context, req *pb.CheckStockRequest) (*pb.CheckStockResponse, error) {
	logger.Debug("gRPC CheckStock called",
		zap.Int64("product_id", req.ProductId),
		zap.Int32("requested_quantity", req.Quantity),
	)

	product, err := h.service.GetByID(ctx, req.ProductId)
	if err != nil {
		logger.Error("gRPC CheckStock failed - product not found",
			zap.Int64("product_id", req.ProductId),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.NotFound, "product not found: %v", err)
	}

	available := int32(product.Stock) >= req.Quantity

	logger.Info("gRPC CheckStock completed",
		zap.Int64("product_id", req.ProductId),
		zap.Int32("current_stock", int32(product.Stock)),
		zap.Int32("requested_quantity", req.Quantity),
		zap.Bool("available", available),
	)

	return &pb.CheckStockResponse{
		Available:    available,
		CurrentStock: int32(product.Stock),
	}, nil
}

// UpdateStock обновляет количество товара на складе
func (h *ProductGRPCHandler) UpdateStock(ctx context.Context, req *pb.UpdateStockRequest) (*pb.UpdateStockResponse, error) {
	logger.Debug("gRPC UpdateStock called",
		zap.Int64("product_id", req.ProductId),
		zap.Int32("quantity_delta", req.QuantityDelta),
	)

	product, err := h.service.GetByID(ctx, req.ProductId)
	if err != nil {
		logger.Error("gRPC UpdateStock failed - product not found",
			zap.Int64("product_id", req.ProductId),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.NotFound, "product not found: %v", err)
	}

	oldStock := int32(product.Stock)
	newStock := oldStock + req.QuantityDelta

	if newStock < 0 {
		logger.Warn("gRPC UpdateStock failed - insufficient stock",
			zap.Int64("product_id", req.ProductId),
			zap.Int32("current_stock", oldStock),
			zap.Int32("requested_delta", req.QuantityDelta),
			zap.Int32("would_be_stock", newStock),
		)
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
		logger.Error("gRPC UpdateStock failed - update error",
			zap.Int64("product_id", req.ProductId),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.Internal, "failed to update stock: %v", err)
	}

	logger.Info("gRPC UpdateStock success",
		zap.Int64("product_id", req.ProductId),
		zap.String("product_name", updatedProduct.Name),
		zap.Int32("old_stock", oldStock),
		zap.Int32("new_stock", int32(updatedProduct.Stock)),
		zap.Int32("delta", req.QuantityDelta),
	)

	return &pb.UpdateStockResponse{
		NewStock: int32(updatedProduct.Stock),
	}, nil
}
