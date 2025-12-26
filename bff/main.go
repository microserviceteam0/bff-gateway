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
	"github.com/microserviceteam0/bff-gateway/bff/internal/handler"
	"github.com/microserviceteam0/bff-gateway/bff/internal/router"
	"github.com/microserviceteam0/bff-gateway/bff/internal/service"
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

	// 2. Инициализация gRPC клиентов
	userAddr := os.Getenv("USER_SERVICE_ADDR")
	if userAddr == "" {
		userAddr = "localhost:50053"
	}
	orderAddr := os.Getenv("ORDER_SERVICE_ADDR")
	if orderAddr == "" {
		orderAddr = "localhost:50051"
	}
	productAddr := os.Getenv("PRODUCT_SERVICE_ADDR")
	if productAddr == "" {
		productAddr = "localhost:50052"
	}

	userConn, err := grpc.NewClient(userAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("failed to connect to user service", "error", err)
		os.Exit(1)
	}
	defer userConn.Close()

	orderConn, err := grpc.NewClient(orderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("failed to connect to order service", "error", err)
		os.Exit(1)
	}
	defer orderConn.Close()

	productConn, err := grpc.NewClient(productAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("failed to connect to product service", "error", err)
		os.Exit(1)
	}
	defer productConn.Close()

	userClient := bffgrpc.NewUserClient(userConn)
	orderClient := bffgrpc.NewOrderClient(orderConn)
	productClient := bffgrpc.NewProductClient(productConn)

	// 3. Инициализация HTTP Clients

	// Auth Service (Token Validation, Login)
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://localhost:8084"
	}
	authClient := clients.NewHTTPAuthClient(authServiceURL)

	// User Service HTTP (Registration)
	userHTTPAddr := os.Getenv("USER_SERVICE_HTTP_ADDR")
	if userHTTPAddr == "" {
		userHTTPAddr = "http://localhost:8081"
	}
	userHTTPClient := clients.NewHTTPUserClient(userHTTPAddr)

	// Product Service HTTP (List Products)
	productHTTPAddr := os.Getenv("PRODUCT_SERVICE_HTTP_ADDR")
	if productHTTPAddr == "" {
		productHTTPAddr = "http://localhost:8082"
	}
	productHTTPClient := clients.NewHTTPProductClient(productHTTPAddr)

	// 4. Инициализация сервисов
	bffService := service.NewBFFService(userClient, orderClient, productClient, authClient, userHTTPClient, productHTTPClient)

	// 5. Инициализация Роутера
	h := handler.NewHandler(bffService)
	r := router.SetupRouter(logger, authClient, h)

	// 6. Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		slog.Info("Server listening", "port", port)
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

