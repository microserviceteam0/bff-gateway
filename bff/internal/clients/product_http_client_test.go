package clients

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestListProducts_RetryAndFallback(t *testing.T) {
	// 1. Success case (after retries or immediate)
	t.Run("Success after retry", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			if attempts < 2 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"id":1, "name":"Product A"}]`))
		}))
		defer server.Close()

		client := NewHTTPProductClient(server.URL, 3, 10*time.Millisecond, 5*time.Second)
		products, err := client.ListProducts(context.Background())
		if err != nil {
			t.Fatalf("expected success, got error: %v", err)
		}
		if len(products) != 1 {
			t.Errorf("expected 1 product, got %d", len(products))
		}
		if attempts != 2 {
			t.Errorf("expected 2 attempts, got %d", attempts)
		}
	})

	// 2. Fallback case (all retries fail)
	t.Run("Fallback on all failures", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewHTTPProductClient(server.URL, 3, 10*time.Millisecond, 5*time.Second)
		products, err := client.ListProducts(context.Background())

		// We expect nil error because of fallback
		if err != nil {
			t.Fatalf("expected nil error (fallback), got: %v", err)
		}
		// We expect empty list
		if len(products) != 0 {
			t.Errorf("expected empty list, got %d items", len(products))
		}
		// Should retry 3 times (initial + 2 retries or whatever retry-go defaults/config is, we set 3 attempts total)
		// We set retry.Attempts(3)
		if attempts != 3 {
			t.Errorf("expected 3 attempts, got %d", attempts)
		}
	})
}
