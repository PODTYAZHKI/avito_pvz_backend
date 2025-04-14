package middleware

import (
	"avito-pvz/internal/metrics"
	"time"

	"github.com/gin-gonic/gin"
)

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()

		c.Next()

		duration := time.Since(start).Seconds()
		status := c.Writer.Status()

		metrics.HttpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
		metrics.HttpRequestsTotal.WithLabelValues(c.Request.Method, path, string(rune(status))).Inc()
	}
}