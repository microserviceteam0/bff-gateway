package service

import (
	pb "order-service/api/order/v1"
	"context"
	"errors"
	"fmt"
	// product_pb "product-service/pkg/api/v1"
	"time"

	"order-service/internal/model"
	"order-service/internal/repository"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderService interface {
	CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error)
	GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error)
	GetUserOrders(ctx context.Context, req *pb.GetUserOrdersRequest) (*pb.GetUserOrdersResponse, error)
	CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error)
	UpdateOrder(ctx context.Context, req *pb.UpdateOrderRequest) (*pb.UpdateOrderResponse, error)
	GetOrderStats(ctx context.Context, req *pb.GetOrderStatsRequest) (*pb.GetOrderStatsResponse, error)
}

type OrderServiceImpl struct {
	repo repository.OrderRepository
	// productClient product_pb.ProductServiceClient
}

func NewOrderService(
	repo repository.OrderRepository,
	// productClient product_pb.ProductServiceClient,
) OrderService {
	return &OrderServiceImpl{
		repo: repo,
		// productClient: productClient,
	}
}

// GRPCServer implements the gRPC server interface for OrderService.
type GRPCServer struct {
	pb.UnimplementedOrderServiceServer // Must be embedded for forward compatibility
	Service                         OrderService       // Our business logic service
}

func (s *GRPCServer) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	return s.Service.CreateOrder(ctx, req)
}

func (s *GRPCServer) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	return s.Service.GetOrder(ctx, req)
}

func (s *GRPCServer) GetUserOrders(ctx context.Context, req *pb.GetUserOrdersRequest) (*pb.GetUserOrdersResponse, error) {
	return s.Service.GetUserOrders(ctx, req)
}

func (s *GRPCServer) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	return s.Service.CancelOrder(ctx, req)
}

func (s *GRPCServer) UpdateOrder(ctx context.Context, req *pb.UpdateOrderRequest) (*pb.UpdateOrderResponse, error) {
	return s.Service.UpdateOrder(ctx, req)
}

func (s *GRPCServer) GetOrderStats(ctx context.Context, req *pb.GetOrderStatsRequest) (*pb.GetOrderStatsResponse, error) {
	return s.Service.GetOrderStats(ctx, req)
}

// CreateOrder создаёт новый заказ
func (s *OrderServiceImpl) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	if err := s.validateCreateOrderRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "INVALID_REQUEST: %v", err)
	}

	// Calls to Product Service were commented out as requested.
	// Using dummy data for product price and name.
	orderItems := make([]model.OrderItem, len(req.Items))
	totalAmount := 0.0
	dummyPrice := 10.0    // 10 units of currency per item
	dummyName := "Sample Product"

	for i, item := range req.Items {
		orderItems[i] = model.OrderItem{
			ProductID:   item.ProductId,
			Quantity:    item.Quantity,
			Price:       dummyPrice, // Using dummy price
			ProductName: fmt.Sprintf("%s %d", dummyName, item.ProductId), // Using dummy name
		}
		totalAmount += dummyPrice * float64(item.Quantity)
	}

	order := &model.Order{
		UserID:      req.UserId,
		Status:      "pending",
		Items:       orderItems,
		TotalAmount: totalAmount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	createdOrder, err := s.repo.CreateOrder(ctx, order)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "DATABASE_ERROR: Failed to create order: %v", err)
	}

	return &pb.CreateOrderResponse{
		Result: &pb.CreateOrderResponse_OrderId{
			OrderId: createdOrder.ID,
		},
	}, nil
}

// GetOrder получает заказ по ID
func (s *OrderServiceImpl) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	userID, isAdmin, err := s.getUserInfoFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "UNAUTHORIZED: %v", err)
	}

	order, err := s.repo.GetOrder(ctx, req.OrderId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "NOT_FOUND: Order not found")
		}
		return nil, status.Errorf(codes.Internal, "DATABASE_ERROR: Failed to get order: %v", err)
	}

	if !isAdmin && order.UserID != userID {
		return nil, status.Errorf(codes.PermissionDenied, "FORBIDDEN: Access denied")
	}

	return &pb.GetOrderResponse{
		Result: &pb.GetOrderResponse_Order{
			Order: s.orderToProto(order),
		},
	}, nil
}

func (s *OrderServiceImpl) GetUserOrders(ctx context.Context, req *pb.GetUserOrdersRequest) (*pb.GetUserOrdersResponse, error) {
	page := req.Page
	pageSize := req.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := int((page - 1) * pageSize)
	limit := int(pageSize)

	orders, err := s.repo.GetOrdersByUserID(ctx, req.UserId, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "DATABASE_ERROR: Failed to get user orders: %v", err)
	}

	// TODO: получить total count из репозитория (добавить в сигнатуру)
	totalCount := int32(len(orders)) // временно

	pbOrders := make([]*pb.Order, len(orders))
	for i, order := range orders {
		pbOrders[i] = s.orderToProto(&order)
	}

	return &pb.GetUserOrdersResponse{
		Orders:     pbOrders,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (s *OrderServiceImpl) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	userID, isAdmin, err := s.getUserInfoFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "UNAUTHORIZED: %v", err)
	}

	order, err := s.repo.GetOrder(ctx, req.OrderId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "NOT_FOUND: Order not found")
		}
		return nil, status.Errorf(codes.Internal, "DATABASE_ERROR: Failed to get order: %v", err)
	}

	if !isAdmin && order.UserID != userID {
		return nil, status.Errorf(codes.PermissionDenied, "FORBIDDEN: Access denied")
	}

	if order.Status != "pending" && order.Status != "confirmed" {
		return nil, status.Errorf(codes.FailedPrecondition, "INVALID_STATUS: Cannot cancel order with status '%s'", order.Status)
	}

	order.Status = "cancelled"
	order.UpdatedAt = time.Now()

	if err := s.repo.UpdateOrder(ctx, order); err != nil {
		return nil, status.Errorf(codes.Internal, "DATABASE_ERROR: Failed to cancel order: %v", err)
	}

	return &pb.CancelOrderResponse{
		Result: &pb.CancelOrderResponse_Success{
			Success: &emptypb.Empty{},
		},
	}, nil
}

func (s *OrderServiceImpl) UpdateOrder(ctx context.Context, req *pb.UpdateOrderRequest) (*pb.UpdateOrderResponse, error) {
	userID, isAdmin, err := s.getUserInfoFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "UNAUTHORIZED: %v", err)
	}

	order, err := s.repo.GetOrder(ctx, req.OrderId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "NOT_FOUND: Order not found")
		}
		return nil, status.Errorf(codes.Internal, "DATABASE_ERROR: Failed to get order: %v", err)
	}

	if !isAdmin && order.UserID != userID {
		return nil, status.Errorf(codes.PermissionDenied, "FORBIDDEN: Access denied")
	}

	if req.Status != nil {
		if !s.isValidStatus(*req.Status) {
			return nil, status.Errorf(codes.InvalidArgument, "INVALID_STATUS: Invalid status")
		}
		order.Status = *req.Status
	}

	order.UpdatedAt = time.Now()

	if err := s.repo.UpdateOrder(ctx, order); err != nil {
		return nil, status.Errorf(codes.Internal, "DATABASE_ERROR: Failed to update order: %v", err)
	}

	return &pb.UpdateOrderResponse{
		Result: &pb.UpdateOrderResponse_Success{
			Success: &emptypb.Empty{},
		},
	}, nil
}

func (s *OrderServiceImpl) GetOrderStats(ctx context.Context, req *pb.GetOrderStatsRequest) (*pb.GetOrderStatsResponse, error) {
	orders, err := s.repo.GetOrdersByUserID(ctx, req.UserId, 1000, 0)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "DATABASE_ERROR: Failed to get order stats: %v", err)
	}

	var totalOrders, activeOrders int32
	var totalSpent float64
	var lastOrderDate *timestamppb.Timestamp

	for _, order := range orders {
		totalOrders++

		if order.Status == "pending" || order.Status == "confirmed" || order.Status == "processing" {
			activeOrders++
		}

		if order.Status == "completed" {
			totalSpent += order.TotalAmount
		}

		if lastOrderDate == nil || order.CreatedAt.After(lastOrderDate.AsTime()) {
			lastOrderDate = timestamppb.New(order.CreatedAt)
		}
	}

	return &pb.GetOrderStatsResponse{
		TotalOrders:   totalOrders,
		ActiveOrders:  activeOrders,
		TotalSpent:    totalSpent,
		LastOrderDate: lastOrderDate,
	}, nil
}

func (s *OrderServiceImpl) validateCreateOrderRequest(req *pb.CreateOrderRequest) error {
	if len(req.Items) == 0 {
		return errors.New("order must contain at least one item")
	}

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return fmt.Errorf("invalid quantity for product %d", item.ProductId)
		}
		if item.ProductId <= 0 {
			return errors.New("invalid product_id")
		}
	}

	return nil
}

/*
func (s *OrderServiceImpl) validateProductAvailability(ctx context.Context, items []*pb.OrderItem) error {
	products := make([]*product_pb.ProductQuantity, len(items))
	for i, item := range items {
		products[i] = &product_pb.ProductQuantity{
			ProductId: item.ProductId,
			Quantity:  item.Quantity,
		}
	}

	resp, err := s.productClient.ValidateProducts(ctx, &product_pb.ValidateProductsRequest{
		Products: products,
	})
	if err != nil {
		return fmt.Errorf("product service error: %w", err)
	}

	if !resp.AllAvailable {
		return s.buildUnavailableError(resp.UnavailableProducts)
	}

	return nil
}

func (s *OrderServiceImpl) fetchProductInfo(ctx context.Context, items []*pb.OrderItem) (map[int64]float64, map[int64]string, error) {
	productIDs := make([]int64, len(items))
	for i, item := range items {
		productIDs[i] = item.ProductId
	}

	resp, err := s.productClient.GetProducts(ctx, &product_pb.GetProductsRequest{
		ProductIds: productIDs,
	})
	if err != nil {
		return nil, nil, err
	}

	prices := make(map[int64]float64)
	names := make(map[int64]string)
	for _, product := range resp.Products {
		prices[product.Id] = product.Price
		names[product.Id] = product.Name
	}

	return prices, names, nil
}

func (s *OrderServiceImpl) buildUnavailableError(unavailable []*product_pb.UnavailableProduct) error {
	msg := "Products unavailable: "
	for i, p := range unavailable {
		if i > 0 {
			msg += ", "
		}
		msg += fmt.Sprintf("product_id=%d (%s)", p.ProductId, p.Reason)
	}
	return errors.New(msg)
}
*/

func (s *OrderServiceImpl) isValidStatus(status string) bool {
	validStatuses := map[string]bool{
		"pending": true, "confirmed": true, "processing": true,
		"completed": true, "cancelled": true,
	}
	return validStatuses[status]
}

func (s *OrderServiceImpl) getUserInfoFromContext(ctx context.Context) (userID int64, isAdmin bool, err error) {
	// TODO: извлечь из gRPC metadata (пробрасывается от Envoy)
	// md, ok := metadata.FromIncomingContext(ctx)
	// userIDStr := md.Get("x-user-id")[0]
	// role := md.Get("x-user-role")[0]
	// return userID, role == "admin", nil

	return 0, false, errors.New("not implemented")
}

func (s *OrderServiceImpl) orderToProto(order *model.Order) *pb.Order {
	items := make([]*pb.OrderItem, len(order.Items))
	for i, item := range order.Items {
		items[i] = &pb.OrderItem{
			ProductId:   item.ProductID,
			Quantity:    item.Quantity,
			Price:       item.Price,
			ProductName: item.ProductName,
		}
	}

	return &pb.Order{
		Id:          order.ID,
		UserId:      order.UserID,
		Status:      order.Status,
		Items:       items,
		TotalAmount: order.TotalAmount,
		CreatedAt:   timestamppb.New(order.CreatedAt),
		UpdatedAt:   timestamppb.New(order.UpdatedAt),
	}
}
