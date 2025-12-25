# Stage 1: Build
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata wget

WORKDIR /build

COPY shared/ ./shared/
COPY product-service/ ./product-service/

WORKDIR /build/product-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/bin/product-service \
    ./cmd/product-service/main.go

# Stage 2: Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget

RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

COPY --from=builder /app/bin/product-service /app/product-service
COPY --from=builder /build/product-service/api/openapi.yaml /app/api/openapi.yaml
COPY --from=builder /build/product-service/migrations /app/migrations

RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8083 50051

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8083/health || exit 1

ENTRYPOINT ["/app/product-service"]
