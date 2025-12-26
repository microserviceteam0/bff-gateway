package grpc

import (
	"context"

	"github.com/microserviceteam0/bff-gateway/bff/internal/clients"
	"google.golang.org/grpc"
	
	productv1 "github.com/microserviceteam0/bff-gateway/bff/api/proto/product"
)

type productClient struct {
	api productv1.ProductServiceClient
}

func NewProductClient(conn *grpc.ClientConn) clients.ProductClient {
	return &productClient{
		api: productv1.NewProductServiceClient(conn),
	}
}

func (c *productClient) GetProduct(ctx context.Context, id int64, opts ...grpc.CallOption) (*productv1.ProductResponse, error) {
	resp, err := c.api.GetProduct(ctx, &productv1.GetProductRequest{Id: id}, opts...)
	return resp, clients.MapGRPCError(err)
}

func (c *productClient) GetProducts(ctx context.Context, ids []int64, opts ...grpc.CallOption) (*productv1.ProductsResponse, error) {
	resp, err := c.api.GetProducts(ctx, &productv1.GetProductsRequest{Ids: ids}, opts...)
	return resp, clients.MapGRPCError(err)
}

func (c *productClient) CheckStock(ctx context.Context, productID int64, quantity int32, opts ...grpc.CallOption) (*productv1.CheckStockResponse, error) {
	resp, err := c.api.CheckStock(ctx, &productv1.CheckStockRequest{
		ProductId: productID,
		Quantity:  quantity,
	}, opts...)
	return resp, clients.MapGRPCError(err)
}

func (c *productClient) UpdateStock(ctx context.Context, productID int64, delta int32, opts ...grpc.CallOption) (*productv1.UpdateStockResponse, error) {
	resp, err := c.api.UpdateStock(ctx, &productv1.UpdateStockRequest{
		ProductId:     productID,
		QuantityDelta: delta,
	}, opts...)
	return resp, clients.MapGRPCError(err)
}