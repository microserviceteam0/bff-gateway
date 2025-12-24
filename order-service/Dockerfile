FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY order-service/ ./order-service/

WORKDIR /build/order-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/orderservice ./cmd/orderservice

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/orderservice .

EXPOSE 50051

ENTRYPOINT ["/app/orderservice"]