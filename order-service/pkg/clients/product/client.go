package product

import (
	"context"
	"fmt"
	"time"

	pb "order-service/pkg/api/product/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn    *grpc.ClientConn
	Service pb.ProductServiceClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to product service: %w", err)
	}

	client := pb.NewProductServiceClient(conn)

	return &Client{
		conn:    conn,
		Service: client,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) GetProducts(ctx context.Context, ids []int64) (map[int64]*pb.ProductResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &pb.GetProductsRequest{
		Ids: ids,
	}

	resp, err := c.Service.GetProducts(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	productMap := make(map[int64]*pb.ProductResponse)
	for _, p := range resp.Products {
		productMap[p.Id] = p
	}

	return productMap, nil
}
