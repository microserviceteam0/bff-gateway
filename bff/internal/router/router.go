package router

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/microserviceteam0/bff-gateway/bff/docs"
	"github.com/microserviceteam0/bff-gateway/bff/internal/clients"
	"github.com/microserviceteam0/bff-gateway/bff/internal/handler"
	"github.com/microserviceteam0/bff-gateway/bff/internal/middleware"
	"github.com/microserviceteam0/bff-gateway/shared/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(
	logger *slog.Logger,
	authClient clients.AuthClient,
	h *handler.Handler,
	rdb *redis.Client,
	cacheTTL time.Duration,
) *gin.Engine {
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(middleware.SlogLogger(logger))
	r.Use(metrics.GinMetricsMiddleware("bff-gateway"))
	r.Use(middleware.RedisCacheMiddleware(rdb, cacheTTL))

	// Metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")

	// Публичные маршруты
	v1.POST("/register", h.Register)
	v1.POST("/login", h.Login)
	v1.GET("/products", h.GetProducts)

	// Защищенные маршруты
	authorized := v1.Group("")
	authorized.Use(middleware.AuthMiddleware(authClient))
	{
		authorized.POST("/orders", h.CreateOrder)
		authorized.GET("/orders/:id", h.GetOrder)
		authorized.POST("/orders/:id/cancel", h.CancelOrder)
		authorized.GET("/profile", h.GetProfile)
	}

	return r
}
