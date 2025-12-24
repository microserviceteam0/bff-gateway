package config

import (
	"os"

	"github.com/joho/godotenv"
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
		DBHost:      getEnv("DB_HOST", "user-db"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "user_user"),
		DBPassword:  getEnv("DB_PASSWORD", "user_password"),
		DBName:      getEnv("DB_NAME", "users_db"),
		AppPort:     getEnv("APP_PORT", "8081"),
		GRPCPort:    getEnv("GRPC_PORT", "50053"), // Значение по умолчанию 50051
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
