FROM golang:1.25-alpine AS builder

WORKDIR /build

# Копируем shared, так как bff зависит от него через replace в go.mod
COPY shared/ ./shared/
COPY bff/ ./bff/

WORKDIR /build/bff

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bff main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bff .

EXPOSE 8080

ENTRYPOINT ["/app/bff"]
