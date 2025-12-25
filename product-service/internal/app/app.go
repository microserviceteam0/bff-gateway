package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/microserviceteam0/bff-gateway/shared/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/microserviceteam0/bff-gateway/product-service/api/proto"
	"github.com/microserviceteam0/bff-gateway/product-service/internal/config"
	"github.com/microserviceteam0/bff-gateway/product-service/internal/handler"
	"github.com/microserviceteam0/bff-gateway/product-service/internal/middleware"
	"github.com/microserviceteam0/bff-gateway/product-service/internal/repository"
	"github.com/microserviceteam0/bff-gateway/product-service/internal/service"
	"github.com/microserviceteam0/bff-gateway/product-service/pkg/database"
	"github.com/microserviceteam0/bff-gateway/product-service/pkg/logger"
)

func Run() error {
	if err := logger.InitLogger("product-service"); err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer func() {
		_ = logger.Log.Sync()
	}()

	logger.Info("starting product service")

	cfg := config.Load()
	logger.Info("configuration loaded",
		zap.String("server_port", cfg.ServerPort),
		zap.String("grpc_port", cfg.GRPCPort),
	)

	db, err := initDatabase(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("failed to init database", zap.Error(err))
		return fmt.Errorf("init database: %w", err)
	}
	defer closeDatabase(db)

	if err := database.RunMigrations(db, "./migrations"); err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
		return fmt.Errorf("run migrations: %w", err)
	}

	productRepo := repository.NewPostgresRepository(db)
	productService := service.NewProductService(productRepo)

	grpcServer := startGRPCServer(cfg.GRPCPort, productService)
	httpServer := startHTTPServer(cfg.ServerPort, productService)

	waitForShutdown(httpServer, grpcServer)
	return nil
}

// initDatabase подключается к PostgreSQL
func initDatabase(databaseURL string) (*sql.DB, error) {
	logger.Info("connecting to database")

	db, err := database.NewPostgres(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	logger.Info("database connected successfully")
	return db, nil
}

// closeDatabase закрывает соединение с БД
func closeDatabase(db *sql.DB) {
	logger.Info("closing database connection")
	if err := db.Close(); err != nil {
		logger.Error("error closing database", zap.Error(err))
	}
}

// startHTTPServer запускает HTTP REST API сервер
func startHTTPServer(port string, productService service.ProductService) *http.Server {
	router := mux.NewRouter()

	router.Use(middleware.LoggingMiddleware)
	router.Use(metrics.HTTPMetricsMiddleware("product-service"))

	productHandler := handler.NewProductHandler(productService)
	productHandler.RegisterRoutes(router)

	router.HandleFunc("/health", healthCheckHandler).Methods(http.MethodGet)
	router.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)
	router.HandleFunc("/api/openapi.yaml", openapiHandler).Methods(http.MethodGet)
	router.PathPrefix("/swagger/").Handler(swaggerHandler(port))

	addr := fmt.Sprintf(":%s", port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logger.Info("HTTP server started",
			zap.String("address", fmt.Sprintf("http://localhost%s", addr)),
			zap.String("api_endpoint", fmt.Sprintf("http://localhost%s/api/products", addr)),
			zap.String("swagger_ui", fmt.Sprintf("http://localhost%s/swagger/", addr)),
		)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("HTTP server failed", zap.Error(err))
		}
	}()

	return server
}

// startGRPCServer запускает gRPC сервер
func startGRPCServer(port string, productService service.ProductService) *grpc.Server {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Fatal("failed to listen gRPC", zap.String("port", port), zap.Error(err))
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(metrics.GRPCUnaryServerInterceptor("product-service")),
	)

	pb.RegisterProductServiceServer(grpcServer, handler.NewProductGRPCHandler(productService))

	go func() {
		logger.Info("gRPC server started",
			zap.String("address", fmt.Sprintf(":%s", port)),
		)

		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("gRPC server failed", zap.Error(err))
		}
	}()

	return grpcServer
}

// healthCheckHandler обрабатывает health check запросы
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// openapiHandler отдаёт OpenAPI спецификацию
func openapiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/yaml")
	http.ServeFile(w, r, "api/openapi.yaml")
}

// swaggerHandler создаёт Swagger UI handler
func swaggerHandler(port string) http.Handler {
	return httpSwagger.Handler(
		httpSwagger.URL("http://localhost:"+port+"/api/openapi.yaml"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("list"),
	)
}

// waitForShutdown ожидает сигнал завершения и gracefully останавливает серверы
func waitForShutdown(httpServer *http.Server, grpcServer *grpc.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info("shutdown signal received, stopping servers")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("HTTP server forced shutdown", zap.Error(err))
	}

	grpcServer.GracefulStop()

	logger.Info("servers stopped gracefully")
}
