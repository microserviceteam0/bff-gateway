package service_test

import (
	"context"
	"errors"
	"order-service/internal/model"
	"order-service/internal/repository"
	"order-service/internal/service"
	"testing"

	pb "order-service/api/order/v1"

	productpb "order-service/pkg/api/product/v1"

	"gorm.io/gorm"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type mockOrderRepository struct {
	createOrderFunc       func(ctx context.Context, order *model.Order) (*model.Order, error)
	getOrderFunc          func(ctx context.Context, orderID int64) (*model.Order, error)
	getOrdersByUserIDFunc func(ctx context.Context, userID int64, limit int, offset int) ([]model.Order, int64, error)
	updateOrderFunc       func(ctx context.Context, order *model.Order) error
	deleteFunc            func(ctx context.Context, orderID int64) error
}

var _ repository.OrderRepository = (*mockOrderRepository)(nil)

func (m *mockOrderRepository) CreateOrder(ctx context.Context, order *model.Order) (*model.Order, error) {
	if m.createOrderFunc != nil {
		return m.createOrderFunc(ctx, order)
	}
	return nil, errors.New("CreateOrder not implemented in mock")
}

func (m *mockOrderRepository) GetOrder(ctx context.Context, orderID int64) (*model.Order, error) {
	if m.getOrderFunc != nil {
		return m.getOrderFunc(ctx, orderID)
	}
	return nil, errors.New("GetOrder not implemented in mock")
}

func (m *mockOrderRepository) GetOrdersByUserID(ctx context.Context, userID int64, limit int, offset int) ([]model.Order, int64, error) {
	if m.getOrdersByUserIDFunc != nil {
		return m.getOrdersByUserIDFunc(ctx, userID, limit, offset)
	}
	return nil, 0, errors.New("GetOrdersByUserID not implemented in mock")
}

func (m *mockOrderRepository) UpdateOrder(ctx context.Context, order *model.Order) error {
	if m.updateOrderFunc != nil {
		return m.updateOrderFunc(ctx, order)
	}
	return errors.New("UpdateOrder not implemented in mock")
}

func (m *mockOrderRepository) Delete(ctx context.Context, orderID int64) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, orderID)
	}
	return errors.New("Delete not implemented in mock")
}

type mockProductClient struct {
	getProductsFunc func(ctx context.Context, ids []int64) (map[int64]*productpb.ProductResponse, error)
}

func (m *mockProductClient) GetProducts(ctx context.Context, ids []int64) (map[int64]*productpb.ProductResponse, error) {
	if m.getProductsFunc != nil {
		return m.getProductsFunc(ctx, ids)
	}
	return nil, errors.New("GetProducts not implemented in mock")
}

// helper to create context with auth metadata
func contextWithAuth(userID string, role string) context.Context {
	md := metadata.New(map[string]string{
		"x-user-id":   userID,
		"x-user-role": role,
	})
	return metadata.NewIncomingContext(context.Background(), md)
}

func TestCreateOrder(t *testing.T) {
	ctx := contextWithAuth("1", "user")

	tests := []struct {
		name            string
		req             *pb.CreateOrderRequest
		mockCreateOrder func(ctx context.Context, order *model.Order) (*model.Order, error)
		mockGetProducts func(ctx context.Context, ids []int64) (map[int64]*productpb.ProductResponse, error)
		expectedCode    codes.Code
		expectedMsg     string
	}{
		{
			name: "Success",
			req: &pb.CreateOrderRequest{
				UserId: 1,
				Items: []*pb.OrderItem{
					{ProductId: 101, Quantity: 2},
				},
			},
			mockGetProducts: func(ctx context.Context, ids []int64) (map[int64]*productpb.ProductResponse, error) {
				return map[int64]*productpb.ProductResponse{
					101: {Id: 101, Name: "Test Product", Price: 10.0},
				}, nil
			},
			mockCreateOrder: func(ctx context.Context, order *model.Order) (*model.Order, error) {
				order.ID = 1
				return order, nil
			},
			expectedCode: codes.OK,
		},
		{
			name: "Forbidden - Create for another user",
			req: &pb.CreateOrderRequest{
				UserId: 2, // Different user
				Items: []*pb.OrderItem{
					{ProductId: 101, Quantity: 1},
				},
			},
			mockGetProducts: nil,
			mockCreateOrder: nil,
			expectedCode:    codes.PermissionDenied,
			expectedMsg:     "FORBIDDEN: Cannot create order for another user",
		},
		{
			name: "Invalid Request - No Items",
			req: &pb.CreateOrderRequest{
				UserId: 1,
				Items:  []*pb.OrderItem{},
			},
			mockGetProducts: nil,
			mockCreateOrder: nil,
			expectedCode:    codes.InvalidArgument,
			expectedMsg:     "INVALID_REQUEST: order must contain at least one item",
		},
		{
			name: "Invalid Request - Zero Quantity",
			req: &pb.CreateOrderRequest{
				UserId: 1,
				Items: []*pb.OrderItem{
					{ProductId: 101, Quantity: 0},
				},
			},
			mockGetProducts: nil,
			mockCreateOrder: nil,
			expectedCode:    codes.InvalidArgument,
			expectedMsg:     "INVALID_REQUEST: invalid quantity for product 101",
		},
		{
			name: "Product Not Found",
			req: &pb.CreateOrderRequest{
				UserId: 1,
				Items: []*pb.OrderItem{
					{ProductId: 999, Quantity: 1},
				},
			},
			mockGetProducts: func(ctx context.Context, ids []int64) (map[int64]*productpb.ProductResponse, error) {
				return map[int64]*productpb.ProductResponse{}, nil // Empty map
			},
			mockCreateOrder: nil,
			expectedCode:    codes.NotFound,
			expectedMsg:     "PRODUCT_NOT_FOUND: Product with ID 999 not found",
		},
		{
			name: "Product Service Error",
			req: &pb.CreateOrderRequest{
				UserId: 1,
				Items: []*pb.OrderItem{
					{ProductId: 101, Quantity: 1},
				},
			},
			mockGetProducts: func(ctx context.Context, ids []int64) (map[int64]*productpb.ProductResponse, error) {
				return nil, errors.New("connection failed")
			},
			mockCreateOrder: nil,
			expectedCode:    codes.Internal,
			expectedMsg:     "PRODUCT_SERVICE_ERROR: Failed to fetch products: connection failed",
		},
		{
			name: "Database Error",
			req: &pb.CreateOrderRequest{
				UserId: 1,
				Items: []*pb.OrderItem{
					{ProductId: 101, Quantity: 1},
				},
			},
			mockGetProducts: func(ctx context.Context, ids []int64) (map[int64]*productpb.ProductResponse, error) {
				return map[int64]*productpb.ProductResponse{
					101: {Id: 101, Name: "Test Product", Price: 10.0},
				}, nil
			},
			mockCreateOrder: func(ctx context.Context, order *model.Order) (*model.Order, error) {
				return nil, errors.New("db connection failed")
			},
			expectedCode: codes.Internal,
			expectedMsg:  "DATABASE_ERROR: Failed to create order: db connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockOrderRepository{
				createOrderFunc: tt.mockCreateOrder,
			}
			mockProd := &mockProductClient{
				getProductsFunc: tt.mockGetProducts,
			}
			s := service.NewOrderService(mockRepo, mockProd)

			resp, err := s.CreateOrder(ctx, tt.req)

			if tt.expectedCode != codes.OK {
				if err == nil {
					t.Fatalf("Expected error, got nil")
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("Expected gRPC status error, got %T", err)
				}
				if st.Code() != tt.expectedCode {
					t.Errorf("Expected status code %v, got %v", tt.expectedCode, st.Code())
				}
				if st.Message() != tt.expectedMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.expectedMsg, st.Message())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
				if resp == nil || resp.GetOrderId() == 0 {
					t.Errorf("Expected order ID, got %v", resp)
				}
			}
		})
	}
}

func TestGetOrder(t *testing.T) {
	testOrder := &model.Order{ID: 1, UserID: 1, Status: "pending", TotalAmount: 50.0}

	tests := []struct {
		name            string
		ctx             context.Context
		req             *pb.GetOrderRequest
		mockGetOrder    func(ctx context.Context, orderID int64) (*model.Order, error)
		expectedCode    codes.Code
		expectedMsg     string
		expectedOrderID int64
	}{
		{
			name:            "Success - Owner Access",
			ctx:             contextWithAuth("1", "user"),
			req:             &pb.GetOrderRequest{OrderId: 1},
			mockGetOrder:    func(ctx context.Context, orderID int64) (*model.Order, error) { return testOrder, nil },
			expectedCode:    codes.OK,
			expectedOrderID: 1,
		},
		{
			name:            "Success - Admin Access",
			ctx:             contextWithAuth("99", "admin"),
			req:             &pb.GetOrderRequest{OrderId: 1},
			mockGetOrder:    func(ctx context.Context, orderID int64) (*model.Order, error) { return testOrder, nil },
			expectedCode:    codes.OK,
			expectedOrderID: 1,
		},
		{
			name:         "Unauthorized - No User Info",
			ctx:          context.Background(),
			req:          &pb.GetOrderRequest{OrderId: 1},
			mockGetOrder: nil,
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "UNAUTHORIZED: metadata is missing",
		},
		{
			name:         "Unauthorized - Invalid User ID Format",
			ctx:          contextWithAuth("invalid-id", "user"),
			req:          &pb.GetOrderRequest{OrderId: 1},
			mockGetOrder: nil,
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "UNAUTHORIZED: invalid x-user-id header: strconv.ParseInt: parsing \"invalid-id\": invalid syntax",
		},
		{
			name:         "Access Denied - Not Owner Nor Admin",
			ctx:          contextWithAuth("2", "user"),
			req:          &pb.GetOrderRequest{OrderId: 1},
			mockGetOrder: func(ctx context.Context, orderID int64) (*model.Order, error) { return testOrder, nil },
			expectedCode: codes.PermissionDenied,
			expectedMsg:  "FORBIDDEN: Access denied",
		},
		{
			name:         "Order Not Found",
			ctx:          contextWithAuth("1", "user"),
			req:          &pb.GetOrderRequest{OrderId: 999},
			mockGetOrder: func(ctx context.Context, orderID int64) (*model.Order, error) { return nil, gorm.ErrRecordNotFound },
			expectedCode: codes.NotFound,
			expectedMsg:  "NOT_FOUND: Order not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockOrderRepository{
				getOrderFunc: tt.mockGetOrder,
			}
			s := service.NewOrderService(mockRepo, nil) // Product client not needed for GetOrder

			resp, err := s.GetOrder(tt.ctx, tt.req)

			if tt.expectedCode != codes.OK {
				if err == nil {
					t.Fatalf("Expected error, got nil")
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("Expected gRPC status error, got %T", err)
				}
				if st.Code() != tt.expectedCode {
					t.Errorf("Expected status code %v, got %v", tt.expectedCode, st.Code())
				}
				if st.Message() != tt.expectedMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.expectedMsg, st.Message())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
				if resp == nil || resp.GetOrder().GetId() != tt.expectedOrderID {
					t.Errorf("Expected order ID %d, got %v", tt.expectedOrderID, resp)
				}
			}
		})
	}
}

func TestGetUserOrders(t *testing.T) {
	ctx := contextWithAuth("1", "user")

	tests := []struct {
		name          string
		ctx           context.Context
		req           *pb.GetUserOrdersRequest
		mockGetOrders func(ctx context.Context, userID int64, limit int, offset int) ([]model.Order, int64, error)
		expectedCode  codes.Code
		expectedMsg   string
		expectedCount int
	}{
		{
			name: "Success",
			ctx:  ctx,
			req:  &pb.GetUserOrdersRequest{UserId: 1, Page: 1, PageSize: 10},
			mockGetOrders: func(ctx context.Context, userID int64, limit int, offset int) ([]model.Order, int64, error) {
				return []model.Order{
					{ID: 1, UserID: 1},
					{ID: 2, UserID: 1},
				}, 2, nil
			},
			expectedCode:  codes.OK,
			expectedCount: 2,
		},
		{
			name:          "Access Denied - Requesting Other User's Orders",
			ctx:           contextWithAuth("2", "user"),
			req:           &pb.GetUserOrdersRequest{UserId: 1},
			mockGetOrders: nil,
			expectedCode:  codes.PermissionDenied,
			expectedMsg:   "FORBIDDEN: Access denied",
		},
		{
			name: "Pagination - Default Values (Page < 1, Size > 100)",
			ctx:  ctx,
			req:  &pb.GetUserOrdersRequest{UserId: 1, Page: 0, PageSize: 1000},
			mockGetOrders: func(ctx context.Context, userID int64, limit int, offset int) ([]model.Order, int64, error) {
				if limit != 20 {
					return nil, 0, errors.New("unexpected limit, expected 20 (default)")
				}
				if offset != 0 {
					return nil, 0, errors.New("unexpected offset, expected 0 (page 1)")
				}
				return []model.Order{}, 0, nil
			},
			expectedCode:  codes.OK,
			expectedCount: 0,
		},
		{
			name: "Database Error",
			ctx:  ctx,
			req:  &pb.GetUserOrdersRequest{UserId: 1, Page: 1, PageSize: 10},
			mockGetOrders: func(ctx context.Context, userID int64, limit int, offset int) ([]model.Order, int64, error) {
				return nil, 0, errors.New("db error")
			},
			expectedCode: codes.Internal,
			expectedMsg:  "DATABASE_ERROR: Failed to get user orders: db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockOrderRepository{
				getOrdersByUserIDFunc: tt.mockGetOrders,
			}
			s := service.NewOrderService(mockRepo, nil)

			resp, err := s.GetUserOrders(tt.ctx, tt.req)

			if tt.expectedCode != codes.OK {
				if err == nil {
					t.Fatalf("Expected error, got nil")
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("Expected gRPC status error, got %T", err)
				}
				if st.Code() != tt.expectedCode {
					t.Errorf("Expected status code %v, got %v", tt.expectedCode, st.Code())
				}
				if st.Message() != tt.expectedMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.expectedMsg, st.Message())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
				if len(resp.GetOrders()) != tt.expectedCount {
					t.Errorf("Expected %d orders, got %d", tt.expectedCount, len(resp.GetOrders()))
				}
			}
		})
	}
}

func TestCancelOrder(t *testing.T) {
	testOrder := &model.Order{ID: 1, UserID: 1, Status: "pending"}

	tests := []struct {
		name            string
		ctx             context.Context
		req             *pb.CancelOrderRequest
		mockGetOrder    func(ctx context.Context, orderID int64) (*model.Order, error)
		mockUpdateOrder func(ctx context.Context, order *model.Order) error
		expectedCode    codes.Code
		expectedMsg     string
	}{
		{
			name:            "Success",
			ctx:             contextWithAuth("1", "user"),
			req:             &pb.CancelOrderRequest{OrderId: 1},
			mockGetOrder:    func(ctx context.Context, orderID int64) (*model.Order, error) { return testOrder, nil },
			mockUpdateOrder: func(ctx context.Context, order *model.Order) error { return nil },
			expectedCode:    codes.OK,
		},
		{
			name:         "Access Denied - Not Owner",
			ctx:          contextWithAuth("2", "user"),
			req:          &pb.CancelOrderRequest{OrderId: 1},
			mockGetOrder: func(ctx context.Context, orderID int64) (*model.Order, error) { return testOrder, nil },
			expectedCode: codes.PermissionDenied,
			expectedMsg:  "FORBIDDEN: Access denied",
		},
		{
			name:         "Order Not Found",
			ctx:          contextWithAuth("1", "user"),
			req:          &pb.CancelOrderRequest{OrderId: 999},
			mockGetOrder: func(ctx context.Context, orderID int64) (*model.Order, error) { return nil, gorm.ErrRecordNotFound },
			expectedCode: codes.NotFound,
			expectedMsg:  "NOT_FOUND: Order not found",
		},
		{
			name: "Invalid Status",
			ctx:  contextWithAuth("1", "user"),
			req:  &pb.CancelOrderRequest{OrderId: 1},
			mockGetOrder: func(ctx context.Context, orderID int64) (*model.Order, error) {
				return &model.Order{ID: 1, UserID: 1, Status: "completed"}, nil
			},
			expectedCode: codes.FailedPrecondition,
			expectedMsg:  "INVALID_STATUS: Cannot cancel order with status 'completed'",
		},
		{
			name: "Database Update Error",
			ctx:  contextWithAuth("1", "user"),
			req:  &pb.CancelOrderRequest{OrderId: 1},
			mockGetOrder: func(ctx context.Context, orderID int64) (*model.Order, error) {
				return &model.Order{ID: 1, UserID: 1, Status: "pending"}, nil
			},
			mockUpdateOrder: func(ctx context.Context, order *model.Order) error { return errors.New("update failed") },
			expectedCode:    codes.Internal,
			expectedMsg:     "DATABASE_ERROR: Failed to cancel order: update failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockOrderRepository{
				getOrderFunc:    tt.mockGetOrder,
				updateOrderFunc: tt.mockUpdateOrder,
			}
			s := service.NewOrderService(mockRepo, nil)

			_, err := s.CancelOrder(tt.ctx, tt.req)

			if tt.expectedCode != codes.OK {
				if err == nil {
					t.Fatalf("Expected error, got nil")
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("Expected gRPC status error, got %T", err)
				}
				if st.Code() != tt.expectedCode {
					t.Errorf("Expected status code %v, got %v", tt.expectedCode, st.Code())
				}
				if st.Message() != tt.expectedMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.expectedMsg, st.Message())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestUpdateOrder(t *testing.T) {
	testOrder := &model.Order{ID: 1, UserID: 1, Status: "pending"}
	newStatus := "confirmed"
	invalidStatus := "weird_status"

	tests := []struct {
		name            string
		ctx             context.Context
		req             *pb.UpdateOrderRequest
		mockGetOrder    func(ctx context.Context, orderID int64) (*model.Order, error)
		mockUpdateOrder func(ctx context.Context, order *model.Order) error
		expectedCode    codes.Code
		expectedMsg     string
	}{
		{
			name:            "Success",
			ctx:             contextWithAuth("1", "user"),
			req:             &pb.UpdateOrderRequest{OrderId: 1, Status: &newStatus},
			mockGetOrder:    func(ctx context.Context, orderID int64) (*model.Order, error) { return testOrder, nil },
			mockUpdateOrder: func(ctx context.Context, order *model.Order) error { return nil },
			expectedCode:    codes.OK,
		},
		{
			name:         "Access Denied - Not Owner",
			ctx:          contextWithAuth("2", "user"),
			req:          &pb.UpdateOrderRequest{OrderId: 1, Status: &newStatus},
			mockGetOrder: func(ctx context.Context, orderID int64) (*model.Order, error) { return testOrder, nil },
			expectedCode: codes.PermissionDenied,
			expectedMsg:  "FORBIDDEN: Access denied",
		},
		{
			name:         "Order Not Found",
			ctx:          contextWithAuth("1", "user"),
			req:          &pb.UpdateOrderRequest{OrderId: 999},
			mockGetOrder: func(ctx context.Context, orderID int64) (*model.Order, error) { return nil, gorm.ErrRecordNotFound },
			expectedCode: codes.NotFound,
			expectedMsg:  "NOT_FOUND: Order not found",
		},
		{
			name:         "Invalid Status",
			ctx:          contextWithAuth("1", "user"),
			req:          &pb.UpdateOrderRequest{OrderId: 1, Status: &invalidStatus},
			mockGetOrder: func(ctx context.Context, orderID int64) (*model.Order, error) { return testOrder, nil },
			expectedCode: codes.InvalidArgument,
			expectedMsg:  "INVALID_STATUS: Invalid status",
		},
		{
			name:            "Database Update Error",
			ctx:             contextWithAuth("1", "user"),
			req:             &pb.UpdateOrderRequest{OrderId: 1, Status: &newStatus},
			mockGetOrder:    func(ctx context.Context, orderID int64) (*model.Order, error) { return testOrder, nil },
			mockUpdateOrder: func(ctx context.Context, order *model.Order) error { return errors.New("update failed") },
			expectedCode:    codes.Internal,
			expectedMsg:     "DATABASE_ERROR: Failed to update order: update failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockOrderRepository{
				getOrderFunc:    tt.mockGetOrder,
				updateOrderFunc: tt.mockUpdateOrder,
			}
			s := service.NewOrderService(mockRepo, nil)

			_, err := s.UpdateOrder(tt.ctx, tt.req)

			if tt.expectedCode != codes.OK {
				if err == nil {
					t.Fatalf("Expected error, got nil")
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("Expected gRPC status error, got %T", err)
				}
				if st.Code() != tt.expectedCode {
					t.Errorf("Expected status code %v, got %v", tt.expectedCode, st.Code())
				}
				if st.Message() != tt.expectedMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.expectedMsg, st.Message())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestGetOrderStats(t *testing.T) {
	ctx := contextWithAuth("1", "user")

	tests := []struct {
		name                string
		ctx                 context.Context
		req                 *pb.GetOrderStatsRequest
		mockGetOrders       func(ctx context.Context, userID int64, limit int, offset int) ([]model.Order, int64, error)
		expectedCode        codes.Code
		expectedMsg         string
		expectedTotalOrders int32
	}{
		{
			name: "Success",
			ctx:  ctx,
			req:  &pb.GetOrderStatsRequest{UserId: 1},
			mockGetOrders: func(ctx context.Context, userID int64, limit int, offset int) ([]model.Order, int64, error) {
				return []model.Order{
					{Status: "completed", TotalAmount: 100},
					{Status: "pending"},
				}, 2, nil
			},
			expectedCode:        codes.OK,
			expectedTotalOrders: 2,
		},
		{
			name:          "Access Denied - Not Owner",
			ctx:           contextWithAuth("2", "user"),
			req:           &pb.GetOrderStatsRequest{UserId: 1},
			mockGetOrders: nil,
			expectedCode:  codes.PermissionDenied,
			expectedMsg:   "FORBIDDEN: Access denied",
		},
		{
			name: "Database Error",
			ctx:  ctx,
			req:  &pb.GetOrderStatsRequest{UserId: 1},
			mockGetOrders: func(ctx context.Context, userID int64, limit int, offset int) ([]model.Order, int64, error) {
				return nil, 0, errors.New("db error")
			},
			expectedCode: codes.Internal,
			expectedMsg:  "DATABASE_ERROR: Failed to get order stats: db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockOrderRepository{
				getOrdersByUserIDFunc: tt.mockGetOrders,
			}
			s := service.NewOrderService(mockRepo, nil)

			resp, err := s.GetOrderStats(tt.ctx, tt.req)

			if tt.expectedCode != codes.OK {
				if err == nil {
					t.Fatalf("Expected error, got nil")
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("Expected gRPC status error, got %T", err)
				}
				if st.Code() != tt.expectedCode {
					t.Errorf("Expected status code %v, got %v", tt.expectedCode, st.Code())
				}
				if st.Message() != tt.expectedMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.expectedMsg, st.Message())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
				if resp.GetTotalOrders() != tt.expectedTotalOrders {
					t.Errorf("Expected %d total orders, got %d", tt.expectedTotalOrders, resp.GetTotalOrders())
				}
			}
		})
	}
}

