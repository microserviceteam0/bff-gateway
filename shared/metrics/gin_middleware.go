package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GinMetricsMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()

		HTTPActiveConnections.WithLabelValues(serviceName).Inc()
		defer HTTPActiveConnections.WithLabelValues(serviceName).Dec()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		HTTPRequestsTotal.WithLabelValues(
			serviceName,
			c.Request.Method,
			c.FullPath(),
			status,
		).Inc()

		HTTPRequestDuration.WithLabelValues(
			serviceName,
			c.Request.Method,
			c.FullPath(),
		).Observe(duration)

		HTTPResponseSize.WithLabelValues(
			serviceName,
			c.FullPath(),
		).Observe(float64(c.Writer.Size()))
	}
}
