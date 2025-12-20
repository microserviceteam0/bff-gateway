package main

import (
	"log/slog"
	"net"
	"os"
	"strconv"
	"time"

	pb "order-service/api/order/v1"
	"order-service/internal/model"
	"order-service/internal/repository"
	"order-service/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	defaultGRPCPort = 50051
	defaultDBDSN    = "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable"
)

func main() {
	// 1. Initialize Logger
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	slog.Info("Starting Order Service...")

	// 2. Load Configuration (GRPC Port, DB DSN)
	grpcPortStr := os.Getenv("GRPC_PORT")
	grpcPort, err := strconv.Atoi(grpcPortStr)
	if err != nil || grpcPort == 0 {
		grpcPort = defaultGRPCPort
		slog.Warn("GRPC_PORT environment variable not set or invalid, using default port", "port", defaultGRPCPort)
	}

	dbDSN := os.Getenv("DATABASE_URL")
	if dbDSN == "" {
		dbDSN = defaultDBDSN // Fallback to a default example DSN
		slog.Warn("DATABASE_URL environment variable not set, using default example DSN. Please configure DATABASE_URL for production.", "dsn", defaultDBDSN)
	}

	// 3. Initialize Database (PostgreSQL)
	db, err := gorm.Open(postgres.Open(dbDSN), &gorm.Config{
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

	// 5. Initialize Service (Product Service client is now removed)
	orderService := service.NewOrderService(orderRepo) // No productClient needed after refactoring
	slog.Info("Order Service initialized.")

	// 6. Setup gRPC Server
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(grpcPort))
	if err != nil {
		slog.Error("Failed to listen", "port", grpcPort, "error", err)
		os.Exit(1)
	}
	slog.Info("gRPC server listening", "address", lis.Addr().String())

	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, &service.GRPCServer{Service: orderService})
	reflection.Register(grpcServer) // Enable gRPC reflection for debugging

	slog.Info("Order Service gRPC server started.")
	if err := grpcServer.Serve(lis); err != nil {
		slog.Error("Failed to serve gRPC server", "error", err)
		os.Exit(1)
	}
}
