package main

import (
	"log/slog"
	"net"
	"order-service/internal/config"
	"order-service/internal/model"
	"order-service/internal/repository"
	"order-service/internal/service"
	"order-service/pkg/clients/product"
	"os"
	"time"

	pb "order-service/api/order/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. Initialize Logger
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	slog.Info("Starting Order Service...")

	// 2. Load Configuration
	cfg := config.Load()

	// 3. Initialize Database (PostgreSQL)
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	slog.Info("Successfully connected to database.")

	// Auto-migrate models
	err = db.AutoMigrate(&model.Order{}, &model.OrderItem{})
	if err != nil {
		slog.Error("Failed to auto-migrate database", "error", err)
		os.Exit(1)
	}
	slog.Info("Database auto-migration completed.")

	// 4. Initialize Repository
	orderRepo := repository.NewOrderRepository(db)
	slog.Info("Order Repository initialized.")

	// 5. Initialize Product Service Client
	productClient, err := product.NewClient(cfg.ProductServiceURL)
	if err != nil {
		slog.Error("Failed to initialize Product Service client", "error", err)
		os.Exit(1)
	}
	defer productClient.Close()
	slog.Info("Product Service client initialized.")

	// 6. Initialize Service
	orderService := service.NewOrderService(orderRepo, productClient)
	slog.Info("Order Service initialized.")

	// 7. Setup gRPC Server
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		slog.Error("Failed to listen", "port", cfg.GRPCPort, "error", err)
		os.Exit(1)
	}
	slog.Info("gRPC server listening", "address", lis.Addr().String())

	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, &service.GRPCServer{Service: orderService})
	reflection.Register(grpcServer)

	slog.Info("Order Service gRPC server started.")
	if err := grpcServer.Serve(lis); err != nil {
		slog.Error("Failed to serve gRPC server", "error", err)
		os.Exit(1)
	}
}
