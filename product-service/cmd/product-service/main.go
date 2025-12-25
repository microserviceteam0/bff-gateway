package main

import (
	"log"

	"github.com/microserviceteam0/bff-gateway/product-service/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("product service failed: %v", err)
	}
}
