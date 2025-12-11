include .env
export

.PHONY: db-start db-stop migrate-up migrate-down migrate-create

# Start PostgresSQL container
db-start:
	docker-compose up -d

# Stop PostgresSQL container
db-stop:
	docker-compose down

# Apply migrations
migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

# Rollback last migration
migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

