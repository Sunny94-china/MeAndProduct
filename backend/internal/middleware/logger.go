package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger 打印 method/path/status/duration
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		if query != "" {
			path = path + "?" + query
		}
		slog.Info(
			"http request",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"size", c.Writer.Size(),
			"duration_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}
