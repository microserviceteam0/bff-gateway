package grpc

import (
	"context"

	"github.com/microserviceteam0/bff-gateway/bff/internal/clients"
	"google.golang.org/grpc"
	
	orderv1 "github.com/microserviceteam0/bff-gateway/bff/api/proto/order/v1"
)

type orderClient struct {
	api orderv1.OrderServiceClient
}

func NewOrderClient(conn *grpc.ClientConn) clients.OrderClient {
	return &orderClient{
		api: orderv1.NewOrderServiceClient(conn),
	}
}

func (c *orderClient) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest, opts ...grpc.CallOption) (*orderv1.CreateOrderResponse, error) {
	resp, err := c.api.CreateOrder(ctx, req, opts...)
	return resp, clients.MapGRPCError(err)
}

func (c *orderClient) CancelOrder(ctx context.Context, orderID, userID int64, reason string, opts ...grpc.CallOption) (*orderv1.CancelOrderResponse, error) {
	resp, err := c.api.CancelOrder(ctx, &orderv1.CancelOrderRequest{
		OrderId: orderID,
		UserId:  userID,
		Reason:  reason,
	}, opts...)
	return resp, clients.MapGRPCError(err)
}

func (c *orderClient) UpdateOrder(ctx context.Context, req *orderv1.UpdateOrderRequest, opts ...grpc.CallOption) (*orderv1.UpdateOrderResponse, error) {
	resp, err := c.api.UpdateOrder(ctx, req, opts...)
	return resp, clients.MapGRPCError(err)
}

func (c *orderClient) GetOrder(ctx context.Context, orderID int64, opts ...grpc.CallOption) (*orderv1.GetOrderResponse, error) {
	resp, err := c.api.GetOrder(ctx, &orderv1.GetOrderRequest{OrderId: orderID}, opts...)
	return resp, clients.MapGRPCError(err)
}

func (c *orderClient) GetUserOrders(ctx context.Context, req *orderv1.GetUserOrdersRequest, opts ...grpc.CallOption) (*orderv1.GetUserOrdersResponse, error) {
	resp, err := c.api.GetUserOrders(ctx, req, opts...)
	return resp, clients.MapGRPCError(err)
}

func (c *orderClient) GetOrderStats(ctx context.Context, userID int64, opts ...grpc.CallOption) (*orderv1.GetOrderStatsResponse, error) {
	resp, err := c.api.GetOrderStats(ctx, &orderv1.GetOrderStatsRequest{UserId: userID}, opts...)
	return resp, clients.MapGRPCError(err)
}