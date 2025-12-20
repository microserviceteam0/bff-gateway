# Stage 1: Builder
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/orderservice ./cmd/orderservice

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/orderservice .

EXPOSE 50051

ENTRYPOINT ["/app/orderservice"]
