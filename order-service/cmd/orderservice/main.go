package main

import (
	"database/sql"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"order-service/internal/config"
	"order-service/internal/middleware"
	"order-service/internal/model"
	"order-service/internal/repository"
	"order-service/internal/service"
	"order-service/pkg/clients/product"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "order-service/api/order/v1"

	"github.com/gin-gonic/gin"
	"github.com/microserviceteam0/bff-gateway/shared/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. Initialize Logger
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
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

	// Monitor DB connection pool
	sqlDB, err := db.DB()
	if err == nil {
		go monitorConnectionPool(sqlDB, "order-service")
	} else {
		slog.Error("Failed to get sql.DB from gorm", "error", err)
	}

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

	// 7. Start Monitoring Server (Gin)
	startMonitoringServer(cfg.MonitoringPort)

	// 8. Setup gRPC Server
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		slog.Error("Failed to listen", "port", cfg.GRPCPort, "error", err)
		os.Exit(1)
	}
	slog.Info("gRPC server listening", "address", lis.Addr().String())

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(metrics.GRPCUnaryServerInterceptor("order-service")),
	)
	pb.RegisterOrderServiceServer(grpcServer, &service.GRPCServer{Service: orderService})
	reflection.Register(grpcServer)

	go func() {
		slog.Info("Order Service gRPC server started.")
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("Failed to serve gRPC server", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutdown signal received, stopping servers...")

	grpcServer.GracefulStop()
	slog.Info("Servers stopped gracefully")
}

func startMonitoringServer(port string) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.LoggingMiddleware())
	router.Use(metrics.GinMetricsMiddleware("order-service"))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		slog.Info("Monitoring server started", "port", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Monitoring server failed", "error", err)
		}
	}()
}

// monitorConnectionPool monitors database connection pool stats
func monitorConnectionPool(db *sql.DB, serviceName string) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := db.Stats()

		metrics.DBConnections.WithLabelValues(serviceName, "open").Set(float64(stats.OpenConnections))
		metrics.DBConnections.WithLabelValues(serviceName, "in_use").Set(float64(stats.InUse))
		metrics.DBConnections.WithLabelValues(serviceName, "idle").Set(float64(stats.Idle))

		slog.Debug("database connection pool stats",
			slog.Int("open_connections", stats.OpenConnections),
			slog.Int("in_use", stats.InUse),
			slog.Int("idle", stats.Idle),
			slog.Int("max_open", stats.MaxOpenConnections),
			slog.Int64("wait_count", stats.WaitCount),
			slog.Duration("wait_duration", stats.WaitDuration),
		)
	}
}
