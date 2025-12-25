# ---------- Stage 1: build ----------
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata wget && \
    update-ca-certificates

WORKDIR /build

COPY shared/ ./shared/
COPY user/ ./user/

WORKDIR /build/user

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/bin/user-service \
    ./cmd/user/main.go


# ---------- Stage 2: runtime ----------
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget && \
    update-ca-certificates

RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

COPY --from=builder /app/bin/user-service /app/user-service
COPY --from=builder /build/user/migrations /app/migrations
COPY --from=builder /build/user/.env.example /app/.env.example

RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8081 50051

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8081/health || exit 1

ENTRYPOINT ["/app/user-service"]
