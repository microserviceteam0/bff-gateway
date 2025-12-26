package clients

import (
	"context"

	"google.golang.org/grpc"

	orderv1 "github.com/microserviceteam0/bff-gateway/bff/api/proto/order/v1"
)

type OrderClient interface {
	CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest, opts ...grpc.CallOption) (*orderv1.CreateOrderResponse, error)
	CancelOrder(ctx context.Context, orderID, userID int64, reason string, opts ...grpc.CallOption) (*orderv1.CancelOrderResponse, error)
	UpdateOrder(ctx context.Context, req *orderv1.UpdateOrderRequest, opts ...grpc.CallOption) (*orderv1.UpdateOrderResponse, error)
	GetOrder(ctx context.Context, orderID int64, opts ...grpc.CallOption) (*orderv1.GetOrderResponse, error)
	GetUserOrders(ctx context.Context, req *orderv1.GetUserOrdersRequest, opts ...grpc.CallOption) (*orderv1.GetUserOrdersResponse, error)
	GetOrderStats(ctx context.Context, userID int64, opts ...grpc.CallOption) (*orderv1.GetOrderStatsResponse, error)
}
