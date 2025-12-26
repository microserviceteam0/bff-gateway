package service

import (
	"context"
	"errors"
	"fmt"
	pb "order-service/api/order/v1"
	"order-service/internal/model"
	"order-service/internal/repository"
	productpb "order-service/pkg/api/product/v1"
	"strconv"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ProductClient interface {
	GetProducts(ctx context.Context, ids []int64) (map[int64]*productpb.ProductResponse, error)
	CheckStock(ctx context.Context, productID int64, quantity int32) (*productpb.CheckStockResponse, error)
	UpdateStock(ctx context.Context, productID int64, quantityDelta int32) (*productpb.UpdateStockResponse, error)
}

type OrderService interface {
	CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error)
	GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error)
	GetUserOrders(ctx context.Context, req *pb.GetUserOrdersRequest) (*pb.GetUserOrdersResponse, error)
	CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error)
	UpdateOrder(ctx context.Context, req *pb.UpdateOrderRequest) (*pb.UpdateOrderResponse, error)
	GetOrderStats(ctx context.Context, req *pb.GetOrderStatsRequest) (*pb.GetOrderStatsResponse, error)
}

type OrderServiceImpl struct {
	repo          repository.OrderRepository
	productClient ProductClient
}

func NewOrderService(
	repo repository.OrderRepository,
	productClient ProductClient,
) OrderService {
	return &OrderServiceImpl{
		repo:          repo,
		productClient: productClient,
	}
}

type GRPCServer struct {
	pb.UnimplementedOrderServiceServer
	Service OrderService
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
	userID, _, err := s.getUserInfoFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "UNAUTHORIZED: %v", err)
	}

	if req.UserId == 0 {
		req.UserId = userID
	} else if req.UserId != userID {
		return nil, status.Errorf(codes.PermissionDenied, "FORBIDDEN: Cannot create order for another user")
	}

	if err := s.validateCreateOrderRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "INVALID_REQUEST: %v", err)
	}

	// 1. Extract product IDs
	productIDs := make([]int64, len(req.Items))
	for i, item := range req.Items {
		productIDs[i] = item.ProductId
	}

	// 2. Fetch product details from Product Service
	productsMap, err := s.productClient.GetProducts(ctx, productIDs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "PRODUCT_SERVICE_ERROR: Failed to fetch products: %v", err)
	}

	orderItems := make([]model.OrderItem, len(req.Items))
	totalAmount := 0.0

	for i, item := range req.Items {
		product, exists := productsMap[item.ProductId]
		if !exists {
			return nil, status.Errorf(codes.NotFound, "PRODUCT_NOT_FOUND: Product with ID %d not found", item.ProductId)
		}

		// Use real price and name from product service
		orderItems[i] = model.OrderItem{
			ProductID:   item.ProductId,
			Quantity:    item.Quantity,
			Price:       product.Price,
			ProductName: product.Name,
		}
		totalAmount += product.Price * float64(item.Quantity)
	}

	// 3. Update stock for each item
	updatedItems := make([]int, 0)
	for i, item := range orderItems {
		_, err := s.productClient.UpdateStock(ctx, item.ProductID, -int32(item.Quantity))
		if err != nil {
			// Rollback previously updated items
			for _, idx := range updatedItems {
				prevItem := orderItems[idx]
				_, _ = s.productClient.UpdateStock(ctx, prevItem.ProductID, int32(prevItem.Quantity))
			}
			return nil, status.Errorf(codes.InvalidArgument, "STOCK_ERROR: Product %d: %v", item.ProductID, err)
		}
		updatedItems = append(updatedItems, i)
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
		for _, idx := range updatedItems {
			item := orderItems[idx]
			_, _ = s.productClient.UpdateStock(ctx, item.ProductID, int32(item.Quantity))
		}
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
	userID, isAdmin, err := s.getUserInfoFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "UNAUTHORIZED: %v", err)
	}

	if !isAdmin && req.UserId != userID {
		return nil, status.Errorf(codes.PermissionDenied, "FORBIDDEN: Access denied")
	}

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

	orders, totalCountRaw, err := s.repo.GetOrdersByUserID(ctx, req.UserId, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "DATABASE_ERROR: Failed to get user orders: %v", err)
	}

	totalCount := int32(totalCountRaw)

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

	for _, item := range order.Items {
		_, err := s.productClient.UpdateStock(ctx, item.ProductID, int32(item.Quantity))
		if err != nil {
			// Log error but don't fail the cancellation as the status is already updated in DB
			fmt.Printf("FAILED_TO_RESTORE_STOCK: product_id=%d, quantity=%d, error=%v\n", item.ProductID, item.Quantity, err)
		}
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

		oldStatus := order.Status
		newStatus := *req.Status

		if oldStatus != "cancelled" && newStatus == "cancelled" {
			for _, item := range order.Items {
				_, err := s.productClient.UpdateStock(ctx, item.ProductID, int32(item.Quantity))
				if err != nil {
					fmt.Printf("FAILED_TO_RESTORE_STOCK: product_id=%d, quantity=%d, error=%v\n", item.ProductID, item.Quantity, err)
				}
			}
		}

		order.Status = newStatus
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
	userID, isAdmin, err := s.getUserInfoFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "UNAUTHORIZED: %v", err)
	}
	if !isAdmin && req.UserId != userID {
		return nil, status.Errorf(codes.PermissionDenied, "FORBIDDEN: Access denied")
	}

	orders, _, err := s.repo.GetOrdersByUserID(ctx, req.UserId, 1000, 0)
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

func (s *OrderServiceImpl) isValidStatus(status string) bool {
	validStatuses := map[string]bool{
		"pending": true, "confirmed": true, "processing": true,
		"completed": true, "cancelled": true,
	}
	return validStatuses[status]
}

func (s *OrderServiceImpl) getUserInfoFromContext(ctx context.Context) (userID int64, isAdmin bool, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, false, errors.New("metadata is missing")
	}

	vals := md.Get("x-user-id")
	if len(vals) == 0 {
		return 0, false, errors.New("x-user-id header is missing")
	}

	id, err := strconv.ParseInt(vals[0], 10, 64)
	if err != nil {
		return 0, false, fmt.Errorf("invalid x-user-id header: %v", err)
	}

	// Check role
	roles := md.Get("x-user-role")
	isAdmin = false
	if len(roles) > 0 && roles[0] == "admin" {
		isAdmin = true
	}

	return id, isAdmin, nil
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
