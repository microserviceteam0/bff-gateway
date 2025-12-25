package database

import (
	"database/sql"
	"fmt"
	"log"
	"user/pkg/config"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func RunMigrations(cfg *config.Config) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to open DB for migrations:", err)
	}
	defer db.Close()

	if err := goose.Up(db, "./migrations"); err != nil {
		log.Fatal("Migration failed:", err)
	}

	log.Println("✅ Migrations applied successfully")
}
