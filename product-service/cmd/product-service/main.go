package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/microserviceteam0/bff-gateway/product-service/internal/config"
	"github.com/microserviceteam0/bff-gateway/product-service/internal/handler"
	"github.com/microserviceteam0/bff-gateway/product-service/internal/repository"
	"github.com/microserviceteam0/bff-gateway/product-service/internal/service"
	"github.com/microserviceteam0/bff-gateway/product-service/pkg/database"
)

func main() {
	cfg := config.Load()
	log.Printf("Starting Product Service...")

	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("Failed to close database connection: %v", err)
		}
	}(db)

	productRepo := repository.NewPostgresRepository(db)
	productService := service.NewProductService(productRepo)
	productHandler := handler.NewProductHandler(productService)

	router := mux.NewRouter()
	productHandler.RegisterRoutes(router)

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}).Methods(http.MethodGet)

	router.HandleFunc("/api/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		http.ServeFile(w, r, "api/openapi.yaml")
	}).Methods(http.MethodGet)

	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("http://localhost:"+cfg.ServerPort+"/api/openapi.yaml"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("list"),
	))

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("✓ HTTP Server starting on http://localhost%s", addr)
	log.Printf("✓ API available at http://localhost%s/api/products", addr)
	log.Printf("✓ OpenAPI spec at http://localhost%s/api/openapi.yaml", addr)
	log.Printf("✓ Swagger UI at http://localhost%s/swagger/", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("HTTP Server failed: %v", err)
	}
}
