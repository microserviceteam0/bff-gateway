// cmd/main.go (обновлённая версия)
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	userv1 "user/api/proto" // исправь путь если нужно
	"user/internal/app/service"
	"user/internal/infrastructure/repository"
	mygrpc "user/internal/transport/grpc" // переименовываем наш пакет
	"user/internal/transport/http/handler"
	"user/internal/transport/middleware"
	"user/pkg/config"
	"user/pkg/database"
	"user/pkg/logger"

	_ "user/api/docs" // Swagger docs

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	grpcgo "google.golang.org/grpc" // переименовываем стандартный gRPC
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title User Service API
// @version 1.0
// @description Микросервис для управления пользователями

// @host localhost:8080
// @BasePath /
func main() {
	// 1. Загружаем конфигурацию
	cfg := config.Load()

	// 2. Инициализируем логгер
	if err := logger.Init(cfg.Environment); err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Log.Sync()

	logger.Log.Info("Starting User Service",
		zap.String("environment", cfg.Environment),
		zap.String("version", "1.0.0"),
	)

	// 3. Выполняем миграции
	database.RunMigrations(cfg)

	// 4. Подключаемся к БД
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Log.Fatal("Failed to connect to database", zap.Error(err))
	}
	logger.Log.Info("Database connection established")

	// 5. Создаём цепочку зависимостей
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)

	// HTTP хэндлеры
	userHandler := handler.NewUserHandler(userService)
	healthHandler := handler.NewHealthHandler(db)

	// gRPC сервер (используем переименованный пакет)
	grpcServer := mygrpc.NewServer(userService) // mygrpc вместо grpc

	// 6. Запускаем gRPC сервер в отдельной горутине
	go startGRPCServer(grpcServer, cfg, logger.Log)

	// 7. Настраиваем Gin для HTTP
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := setupHTTPRouter(userHandler, healthHandler, logger.Log)

	// 8. Запускаем HTTP сервер с graceful shutdown
	startHTTPServer(router, cfg, logger.Log)
}

// startGRPCServer запускает gRPC сервер
func startGRPCServer(server *mygrpc.Server, cfg *config.Config, logger *zap.Logger) {
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		logger.Fatal("Failed to listen for gRPC",
			zap.String("port", cfg.GRPCPort),
			zap.Error(err),
		)
	}

	s := grpcgo.NewServer() // grpcgo вместо grpc
	userv1.RegisterUserServiceServer(s, server)

	// Включаем reflection для тестирования (dev only)
	if cfg.Environment != "production" {
		reflection.Register(s)
	}

	logger.Info("gRPC server listening",
		zap.String("port", cfg.GRPCPort),
		zap.String("address", "0.0.0.0:"+cfg.GRPCPort),
	)

	if err := s.Serve(lis); err != nil {
		logger.Fatal("Failed to serve gRPC", zap.Error(err))
	}
}

// setupHTTPRouter настраивает HTTP роутер
func setupHTTPRouter(userHandler *handler.UserHandler, healthHandler *handler.HealthHandler, logger *zap.Logger) *gin.Engine {
	router := gin.New()

	// Middleware (важен порядок!)
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.RecoveryMiddleware(logger))

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	router.GET("/health", healthHandler.HealthCheck)

	// API routes
	api := router.Group("/api")
	{
		users := api.Group("/users")
		{
			users.POST("", userHandler.CreateUser)
			users.GET("", userHandler.GetAllUsers)
			users.GET("/:id", userHandler.GetUser)
			users.GET("/email/:email", userHandler.GetUserByEmail)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
			users.POST("/validate", userHandler.ValidateUser)
		}
	}

	// Root endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service":   "User Service API",
			"version":   "1.0.0",
			"status":    "running",
			"protocols": []string{"HTTP/REST", "gRPC"},
			"endpoints": gin.H{
				"docs":   "/swagger/index.html",
				"health": "/health",
				"api":    "/api/users",
			},
		})
	})

	return router
}

// startHTTPServer запускает HTTP сервер с graceful shutdown
func startHTTPServer(router *gin.Engine, cfg *config.Config, logger *zap.Logger) {
	// Создаём HTTP сервер с таймаутами
	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запускаем сервер в отдельной горутине
	go func() {
		logger.Info("HTTP server listening",
			zap.String("port", cfg.AppPort),
			zap.String("address", "0.0.0.0:"+cfg.AppPort),
			zap.String("swagger", "http://localhost:"+cfg.AppPort+"/swagger/index.html"),
		)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exiting")
}
