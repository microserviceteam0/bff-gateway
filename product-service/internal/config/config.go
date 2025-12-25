package config

import (
	"os"
)

type Config struct {
	DatabaseURL string
	ServerPort  string
	GRPCPort    string
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://product_user:product_password@localhost:5433/products_db?sslmode=disable"),
		ServerPort:  getEnv("PRODUCT_SERVER_PORT", "8083"),
		GRPCPort:    getEnv("PRODUCT_GRPC_PORT", "50051"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
