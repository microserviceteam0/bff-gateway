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
	"user/internal/domain/repository"

	userv1 "user/api/proto"
	"user/internal/app/user_service"
	mygrpc "user/internal/transport/grpc"
	"user/internal/transport/http/handler"
	"user/internal/transport/middleware"
	"user/pkg/config"
	"user/pkg/database"
	"user/pkg/logger"

	_ "user/api/docs"

	"github.com/gin-gonic/gin"
	"github.com/microserviceteam0/bff-gateway/shared/metrics"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	grpcgo "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {

	cfg := config.Load()

	if err := logger.Init(cfg.Environment); err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Log.Sync()

	logger.Log.Info("Starting User Service",
		zap.String("environment", cfg.Environment),
		zap.String("version", "1.0.0"),
	)

	database.RunMigrations(cfg)

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Log.Fatal("Failed to connect to database", zap.Error(err))
	}
	logger.Log.Info("Database connection established")

	userRepo := repository.NewUserRepository(db)
	userService := user_service.NewUserService(userRepo)

	userHandler := handler.NewUserHandler(userService)
	healthHandler := handler.NewHealthHandler(db)

	grpcServer := mygrpc.NewServer(userService)

	go startGRPCServer(grpcServer, cfg, logger.Log)

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := setupHTTPRouter(userHandler, healthHandler, logger.Log)

	startHTTPServer(router, cfg, logger.Log)
}

func startGRPCServer(server *mygrpc.Server, cfg *config.Config, logger *zap.Logger) {
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		logger.Fatal("Failed to listen for gRPC",
			zap.String("port", cfg.GRPCPort),
			zap.Error(err),
		)
	}

	s := grpcgo.NewServer()
	userv1.RegisterUserServiceServer(s, server)

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

func setupHTTPRouter(userHandler *handler.UserHandler, healthHandler *handler.HealthHandler, logger *zap.Logger) *gin.Engine {
	router := gin.New()

	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(metrics.GinMetricsMiddleware("user-service"))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/health", healthHandler.HealthCheck)

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

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"auth_service": "User Service API",
			"version":      "1.0.0",
			"status":       "running",
			"protocols":    []string{"HTTP/REST", "gRPC"},
			"endpoints": gin.H{
				"docs":   "/swagger/index.html",
				"health": "/health",
				"api":    "/api/users",
			},
		})
	})

	return router
}

func startHTTPServer(router *gin.Engine, cfg *config.Config, logger *zap.Logger) {
	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

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
