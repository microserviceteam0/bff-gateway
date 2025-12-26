package clients

import (
	"context"

	"google.golang.org/grpc"

	productv1 "github.com/microserviceteam0/bff-gateway/bff/api/proto/product"
)

type ProductClient interface {
	GetProduct(ctx context.Context, id int64, opts ...grpc.CallOption) (*productv1.ProductResponse, error)
	GetProducts(ctx context.Context, ids []int64, opts ...grpc.CallOption) (*productv1.ProductsResponse, error)
	CheckStock(ctx context.Context, productID int64, quantity int32, opts ...grpc.CallOption) (*productv1.CheckStockResponse, error)
	UpdateStock(ctx context.Context, productID int64, delta int32, opts ...grpc.CallOption) (*productv1.UpdateStockResponse, error)
}
