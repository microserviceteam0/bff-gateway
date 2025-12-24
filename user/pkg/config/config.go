package config

import (
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	AppPort     string
	GRPCPort    string // Добавляем GRPC порт
	Environment string
}

func Load() *Config {
	godotenv.Load()

	config := &Config{
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", "postgres"),
		DBName:      getEnv("DB_NAME", "user_service"),
		AppPort:     getEnv("APP_PORT", "8080"),
		GRPCPort:    getEnv("GRPC_PORT", "50051"), // Значение по умолчанию 50051
		Environment: getEnv("ENVIRONMENT", "development"),
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
