package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"user/internal/app/auth_service"
	"user/internal/transport/http/handler"

	"user/internal/app/user_service"
	"user/internal/domain/repository"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type UserAdapter struct {
	userService user_service.UserService
}

func (a *UserAdapter) ValidateCredentials(email, password string) (*struct {
	ID    int64
	Email string
	Role  string
}, error) {

	user, err := a.userService.ValidateCredentials(email, password)
	if err != nil {
		return nil, err
	}

	return &struct {
		ID    int64
		Email string
		Role  string
	}{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}

func main() {

	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env.dev file not found")
	}

	db, err := initDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	userService := user_service.NewUserService(userRepo)

	userAdapter := &UserAdapter{userService: userService}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production"
	}

	tokenDuration := 24 * time.Hour
	if durStr := os.Getenv("TOKEN_DURATION_HOURS"); durStr != "" {
		if hours, err := time.ParseDuration(durStr + "h"); err == nil {
			tokenDuration = hours
		}
	}

	authService := auth_service.NewAuthService(userAdapter, jwtSecret, tokenDuration)

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {

		dbStatus := "healthy"
		if db != nil {
			sqlDB, err := db.DB()
			if err != nil {
				dbStatus = "unhealthy"
			} else if err := sqlDB.Ping(); err != nil {
				dbStatus = "unhealthy"
			}
		}

		status := 200
		if dbStatus != "healthy" {
			status = 503
		}

		c.JSON(status, gin.H{
			"status":       dbStatus,
			"user_service": "auth-user_service",
			"timestamp":    time.Now().Format(time.RFC3339),
			"version":      "1.0.0",
			"database":     dbStatus,
		})
	})

	authHandler := handler.NewAuthHandler(authService)
	authHandler.RegisterRoutes(r)

	port := os.Getenv("AUTH_PORT")
	if port == "" {
		port = "8082"
	}

	log.Printf("Auth user_service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
