package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port               string
	UserServiceAddr    string
	OrderServiceAddr   string
	ProductServiceAddr string
	AuthServiceURL     string
	UserServiceHTTP    string
	ProductServiceHTTP string
	RedisAddr          string
	CacheTTL           time.Duration
	
	// Rate Limit
	RateLimitRPS   float64
	RateLimitBurst int

	// Retry / Resilience
	RetryAttempts uint
	RetryDelay    time.Duration

	// Timeouts
	HttpClientTimeout time.Duration
	ShutdownTimeout   time.Duration
}

func Load() *Config {
	return &Config{
		Port:               getEnv("PORT", "8080"),
		UserServiceAddr:    getEnv("USER_SERVICE_ADDR", "localhost:50053"),
		OrderServiceAddr:   getEnv("ORDER_SERVICE_ADDR", "localhost:50051"),
		ProductServiceAddr: getEnv("PRODUCT_SERVICE_ADDR", "localhost:50052"),
		AuthServiceURL:     getEnv("AUTH_SERVICE_URL", "http://localhost:8084"),
		UserServiceHTTP:    getEnv("USER_SERVICE_HTTP_ADDR", "http://localhost:8081"),
		ProductServiceHTTP: getEnv("PRODUCT_SERVICE_HTTP_ADDR", "http://localhost:8082"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		CacheTTL:           getEnvDuration("CACHE_TTL_SECONDS", 30) * time.Second,
		
		RateLimitRPS:   getEnvFloat("RATE_LIMIT_RPS", 10.0),
		RateLimitBurst: getEnvInt("RATE_LIMIT_BURST", 20),
		RetryAttempts:  uint(getEnvInt("RETRY_ATTEMPTS", 3)),
		RetryDelay:     getEnvDuration("RETRY_DELAY_MS", 200) * time.Millisecond,

		HttpClientTimeout: getEnvDuration("HTTP_CLIENT_TIMEOUT_MS", 5000) * time.Millisecond,
		ShutdownTimeout:   getEnvDuration("SHUTDOWN_TIMEOUT_SECONDS", 5) * time.Second,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}
	return val
}

func getEnvFloat(key string, defaultValue float64) float64 {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return defaultValue
	}
	return val
}

func getEnvDuration(key string, defaultValue int) time.Duration {
	valStr := os.Getenv(key)
	if valStr == "" {
		return time.Duration(defaultValue)
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return time.Duration(defaultValue)
	}
	return time.Duration(val)
}
