package config

import (
	"os"
)

type Config struct {
	GRPCPort          string
	DatabaseURL       string
	ProductServiceURL string
}

func Load() *Config {
	return &Config{
		GRPCPort:          getEnv("GRPC_PORT", "50051"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://order_user:order_password@localhost:5434/orders_db?sslmode=disable"),
		ProductServiceURL: getEnv("PRODUCT_SERVICE_URL", "localhost:50051"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
