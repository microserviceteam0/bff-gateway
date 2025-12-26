package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/avast/retry-go/v4"
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
	baseURL       string
	httpClient    *http.Client
	retryAttempts uint
	retryDelay    time.Duration
}

func NewHTTPProductClient(baseURL string, retryAttempts uint, retryDelay time.Duration, timeout time.Duration) ProductHTTPClient {
	return &httpProductClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		retryAttempts: retryAttempts,
		retryDelay:    retryDelay,
	}
}

func (c *httpProductClient) ListProducts(ctx context.Context) ([]ProductHTTPResponse, error) {
	var products []ProductHTTPResponse

	err := retry.Do(
		func() error {
			req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/products", nil)
			if err != nil {
				return retry.Unrecoverable(fmt.Errorf("failed to create request: %w", err))
			}

			resp, err := c.httpClient.Do(req)
			if err != nil {
				return err // Retry on network errors
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				// Decide if you want to retry on 5xx or not. usually 5xx are retryable, 4xx are not.
				if resp.StatusCode >= 500 {
					return fmt.Errorf("server error: %d", resp.StatusCode)
				}
				return retry.Unrecoverable(MapStatusToError(resp.StatusCode, "failed to list products"))
			}

			if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
				return retry.Unrecoverable(fmt.Errorf("failed to decode response: %w", err))
			}

			return nil
		},
		retry.Context(ctx),
		retry.Attempts(c.retryAttempts),
		retry.Delay(c.retryDelay),
		retry.DelayType(retry.BackOffDelay),
		retry.OnRetry(func(n uint, err error) {
			slog.Warn("Retrying request", "attempt", n+1, "error", err)
		}),
	)

	if err != nil {
		// FALLBACK: Return empty list instead of error
		slog.Error("All retries failed for ListProducts. Falling back to empty list", "error", err)
		return []ProductHTTPResponse{}, nil
	}

	return products, nil
}
