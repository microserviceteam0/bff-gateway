package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"

	pb "github.com/microserviceteam0/bff-gateway/product-service/api/proto"
	"github.com/microserviceteam0/bff-gateway/product-service/internal/config"
	"github.com/microserviceteam0/bff-gateway/product-service/internal/handler"
	"github.com/microserviceteam0/bff-gateway/product-service/internal/repository"
	"github.com/microserviceteam0/bff-gateway/product-service/internal/service"
	"github.com/microserviceteam0/bff-gateway/product-service/pkg/database"
)

func Run() error {
	cfg := config.Load()
	log.Println("Starting Product Service...")

	db, err := initDatabase(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("init database: %w", err)
	}
	defer closeDatabase(db)

	productRepo := repository.NewPostgresRepository(db)
	productService := service.NewProductService(productRepo)

	grpcServer := startGRPCServer(cfg.GRPCPort, productService)
	httpServer := startHTTPServer(cfg.ServerPort, productService)

	waitForShutdown(httpServer, grpcServer)
	return nil
}

// initDatabase подключается к PostgreSQL
func initDatabase(databaseURL string) (*sql.DB, error) {
	db, err := database.NewPostgres(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	log.Println("✓ Database connected")
	return db, nil
}

// closeDatabase закрывает соединение с БД
func closeDatabase(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}
}

// startHTTPServer запускает HTTP REST API сервер
func startHTTPServer(port string, productService service.ProductService) *http.Server {
	router := mux.NewRouter()

	productHandler := handler.NewProductHandler(productService)
	productHandler.RegisterRoutes(router)

	router.HandleFunc("/health", healthCheckHandler).Methods(http.MethodGet)

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
		log.Printf("✓ HTTP Server starting on http://localhost%s", addr)
		log.Printf("✓ API available at http://localhost%s/api/products", addr)
		log.Printf("✓ Swagger UI at http://localhost%s/swagger/", addr)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP Server failed: %v", err)
		}
	}()

	return server
}

// startGRPCServer запускает gRPC сервер
func startGRPCServer(port string, productService service.ProductService) *grpc.Server {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen gRPC: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterProductServiceServer(grpcServer, handler.NewProductGRPCHandler(productService))

	go func() {
		log.Printf("✓ gRPC Server starting on :%s", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC Server failed: %v", err)
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
	log.Println("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP Server forced to shutdown: %v", err)
	}

	grpcServer.GracefulStop()

	log.Println("Servers stopped")
}
