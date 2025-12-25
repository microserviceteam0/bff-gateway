# =========================
# Stage 1: Build
# =========================
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata && \
    update-ca-certificates

WORKDIR /app

# Зависимости
COPY go.mod go.sum ./
RUN go mod download

# Исходники
COPY . .

# Сборка Auth Service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" \
    -o /app/auth-service \
    ./cmd/user_auth

# =========================
# Stage 2: Runtime
# =========================
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata && \
    update-ca-certificates

# Non-root пользователь
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Бинарник
COPY --from=builder /app/auth-service .

# ENV (если используется)
COPY --from=builder /app/.env.dev ./.env

RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8081

CMD ["./auth-service"]
