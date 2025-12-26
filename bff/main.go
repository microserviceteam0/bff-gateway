package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/microserviceteam0/bff-gateway/bff/internal/clients"
	bffgrpc "github.com/microserviceteam0/bff-gateway/bff/internal/clients/grpc"
	"github.com/microserviceteam0/bff-gateway/bff/internal/config"
	"github.com/microserviceteam0/bff-gateway/bff/internal/handler"
	"github.com/microserviceteam0/bff-gateway/bff/internal/router"
	"github.com/microserviceteam0/bff-gateway/bff/internal/service"
	"github.com/microserviceteam0/bff-gateway/shared/metrics"
	"github.com/redis/go-redis/v9"
)

// @title           BFF Gateway API
// @version         1.0
// @description     This is the BFF Gateway for the Microservices E-commerce App.
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// 1. Инициализация логгера
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Starting BFF Gateway...")

	// 2. Загрузка конфига
	cfg := config.Load()

	// 3. Инициализация Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	// Проверка соединения
	ctxPing, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelPing()
	if err := rdb.Ping(ctxPing).Err(); err != nil {
		slog.Warn("Failed to connect to Redis, caching will not work", "error", err)
	} else {
		slog.Info("Connected to Redis", "addr", cfg.RedisAddr)
		go monitorRedisPool(rdb, "bff-gateway")
	}

	// 4. Инициализация gRPC клиентов
	userConn, err := grpc.NewClient(cfg.UserServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("failed to connect to user service", "error", err)
		os.Exit(1)
	}
	defer userConn.Close()

	orderConn, err := grpc.NewClient(cfg.OrderServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("failed to connect to order service", "error", err)
		os.Exit(1)
	}
	defer orderConn.Close()

	productConn, err := grpc.NewClient(cfg.ProductServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("failed to connect to product service", "error", err)
		os.Exit(1)
	}
	defer productConn.Close()

	userClient := bffgrpc.NewUserClient(userConn)
	orderClient := bffgrpc.NewOrderClient(orderConn)
	productClient := bffgrpc.NewProductClient(productConn)

	// 5. Инициализация HTTP Clients
	authClient := clients.NewHTTPAuthClient(cfg.AuthServiceURL)
	userHTTPClient := clients.NewHTTPUserClient(cfg.UserServiceHTTP)
	productHTTPClient := clients.NewHTTPProductClient(cfg.ProductServiceHTTP)

	// 6. Инициализация сервисов
	bffService := service.NewBFFService(userClient, orderClient, productClient, authClient, userHTTPClient, productHTTPClient)

	// 7. Инициализация Роутера
	h := handler.NewHandler(bffService)
	r := router.SetupRouter(logger, authClient, h, rdb, cfg.CacheTTL)

	// 8. Запуск сервера
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		slog.Info("Server listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown: ", err)
	}

	slog.Info("Server exiting")
}

func monitorRedisPool(rdb *redis.Client, serviceName string) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := rdb.PoolStats()

		metrics.DBConnections.WithLabelValues(serviceName, "open").Set(float64(stats.TotalConns))
		metrics.DBConnections.WithLabelValues(serviceName, "in_use").Set(float64(stats.TotalConns - stats.IdleConns))
		metrics.DBConnections.WithLabelValues(serviceName, "idle").Set(float64(stats.IdleConns))

		slog.Debug("redis connection pool stats",
			slog.Uint64("total_connections", uint64(stats.TotalConns)),
			slog.Uint64("idle_connections", uint64(stats.IdleConns)),
			slog.Uint64("stale_connections", uint64(stats.StaleConns)),
		)
	}
}
