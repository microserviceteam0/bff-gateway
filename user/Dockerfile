# Dockerfile
# Этап 1: Сборка
FROM golang:1.21-alpine AS builder

# Устанавливаем зависимости для компиляции
RUN apk add --no-cache git ca-certificates tzdata && \
    update-ca-certificates

# Создаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /app/user-auth_service ./cmd/main.go

# Этап 2: Запуск
FROM alpine:latest

# Устанавливаем зависимости для runtime
RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates

# Создаем не-root пользователя
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Создаем рабочую директорию
WORKDIR /app

# Копируем бинарник из builder этапа
COPY --from=builder /app/user-service .

# Копируем миграции
COPY --from=builder /app/migrations ./migrations

# Копируем конфигурационные файлы (если есть)
COPY --from=builder /app/.env.example ./

# Меняем владельца файлов
RUN chown -R appuser:appuser /app

# Переключаемся на не-root пользователя
USER appuser

# Экспортируем порт
EXPOSE 8080

# Команда запуска
CMD ["./user-service"]