package test

import (
	"context"
	"log"
	"testing"
	"time"

	pb "github.com/microserviceteam0/bff-gateway/product-service/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestGRPCClient(t *testing.T) {
	conn, err := grpc.NewClient("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	client := pb.NewProductServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Тест GetProduct
	t.Run("GetProduct", func(t *testing.T) {
		resp, err := client.GetProduct(ctx, &pb.GetProductRequest{Id: 1})
		if err != nil {
			t.Errorf("GetProduct failed: %v", err)
			return
		}
		log.Printf("Product: ID=%d, Name=%s, Price=%.2f, Stock=%d",
			resp.Id, resp.Name, resp.Price, resp.Stock)

		if resp.Id != 1 {
			t.Errorf("Expected ID=1, got %d", resp.Id)
		}
	})

	// Тест CheckStock
	t.Run("CheckStock", func(t *testing.T) {
		resp, err := client.CheckStock(ctx, &pb.CheckStockRequest{
			ProductId: 1,
			Quantity:  5,
		})
		if err != nil {
			t.Errorf("CheckStock failed: %v", err)
			return
		}
		log.Printf("Stock check: Available=%v, CurrentStock=%d",
			resp.Available, resp.CurrentStock)
	})

	// Тест GetProducts
	t.Run("GetProducts", func(t *testing.T) {
		resp, err := client.GetProducts(ctx, &pb.GetProductsRequest{
			Ids: []int64{1, 2, 3},
		})
		if err != nil {
			t.Errorf("GetProducts failed: %v", err)
			return
		}
		log.Printf("Products count: %d", len(resp.Products))
		for _, p := range resp.Products {
			log.Printf("  - %s (ID: %d, Price: %.2f)", p.Name, p.Id, p.Price)
		}

		if len(resp.Products) == 0 {
			t.Error("Expected at least 1 product")
		}
	})

	// Тест UpdateStock (уменьшаем на 2, потом возвращаем обратно)
	t.Run("UpdateStock", func(t *testing.T) {
		resp1, err := client.UpdateStock(ctx, &pb.UpdateStockRequest{
			ProductId:     1,
			QuantityDelta: -2,
		})
		if err != nil {
			t.Errorf("UpdateStock (decrease) failed: %v", err)
			return
		}
		log.Printf("Stock decreased: NewStock=%d", resp1.NewStock)

		resp2, err := client.UpdateStock(ctx, &pb.UpdateStockRequest{
			ProductId:     1,
			QuantityDelta: 2,
		})
		if err != nil {
			t.Errorf("UpdateStock (increase) failed: %v", err)
			return
		}
		log.Printf("Stock restored: NewStock=%d", resp2.NewStock)
	})
}
