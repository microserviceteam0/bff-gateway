package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"

	"github.com/microserviceteam0/bff-gateway/product-service/pkg/logger"
)

func RunMigrations(db *sql.DB, migrationsPath string) error {
	logger.Info("running database migrations",
		zap.String("path", migrationsPath),
	)

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("get migration version: %w", err)
	}

	if errors.Is(err, migrate.ErrNilVersion) {
		logger.Info("no migrations applied yet")
	} else {
		logger.Info("migrations completed successfully",
			zap.Uint("version", version),
			zap.Bool("dirty", dirty),
		)
	}

	return nil
}
