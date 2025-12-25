package metrics

import (
	"net/http"
	"strconv"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += int64(n)
	return n, err
}

func HTTPMetricsMiddleware(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()

			HTTPActiveConnections.WithLabelValues(serviceName).Inc()
			defer HTTPActiveConnections.WithLabelValues(serviceName).Dec()

			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start).Seconds()
			status := strconv.Itoa(wrapped.statusCode)

			HTTPRequestsTotal.WithLabelValues(
				serviceName,
				r.Method,
				r.URL.Path,
				status,
			).Inc()

			HTTPRequestDuration.WithLabelValues(
				serviceName,
				r.Method,
				r.URL.Path,
			).Observe(duration)

			HTTPResponseSize.WithLabelValues(
				serviceName,
				r.URL.Path,
			).Observe(float64(wrapped.size))
		})
	}
}
