package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ProductHTTPResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int32   `json:"stock"`
}

type ProductHTTPClient interface {
	ListProducts(ctx context.Context) ([]ProductHTTPResponse, error)
}

type httpProductClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPProductClient(baseURL string) ProductHTTPClient {
	return &httpProductClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *httpProductClient) ListProducts(ctx context.Context) ([]ProductHTTPResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/products", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("product service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, MapStatusToError(resp.StatusCode, "failed to list products")
	}

	var products []ProductHTTPResponse
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return products, nil
}
