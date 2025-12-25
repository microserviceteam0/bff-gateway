package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/microserviceteam0/bff-gateway/shared/metrics"

	"github.com/microserviceteam0/bff-gateway/product-service/pkg/logger"
	"go.uber.org/zap"
)

func NewPostgres(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("database connection pool configured",
		zap.Int("max_open_conns", 25),
		zap.Int("max_idle_conns", 5),
		zap.Duration("conn_max_lifetime", 5*time.Minute),
	)

	go monitorConnectionPool(db, "product-service")

	return db, nil
}

// monitorConnectionPool отслеживает состояние connection pool и отправляет метрики в Prometheus
func monitorConnectionPool(db *sql.DB, serviceName string) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := db.Stats()

		metrics.DBConnections.WithLabelValues(serviceName, "open").Set(float64(stats.OpenConnections))
		metrics.DBConnections.WithLabelValues(serviceName, "in_use").Set(float64(stats.InUse))
		metrics.DBConnections.WithLabelValues(serviceName, "idle").Set(float64(stats.Idle))

		logger.Debug("database connection pool stats",
			zap.Int("open_connections", stats.OpenConnections),
			zap.Int("in_use", stats.InUse),
			zap.Int("idle", stats.Idle),
			zap.Int("max_open", stats.MaxOpenConnections),
			zap.Int64("wait_count", stats.WaitCount),
			zap.Duration("wait_duration", stats.WaitDuration),
			zap.Int64("max_idle_closed", stats.MaxIdleClosed),
			zap.Int64("max_lifetime_closed", stats.MaxLifetimeClosed),
		)

		if stats.WaitCount > 100 {
			logger.Warn("high database connection wait count",
				zap.Int64("wait_count", stats.WaitCount),
				zap.Duration("wait_duration", stats.WaitDuration),
			)
		}

		if stats.InUse == stats.MaxOpenConnections {
			logger.Warn("all database connections are in use",
				zap.Int("in_use", stats.InUse),
				zap.Int("max_open", stats.MaxOpenConnections),
			)
		}
	}
}
