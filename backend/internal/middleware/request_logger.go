package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDKey = "request_id"

// RequestID injects a unique X-Request-ID into each request.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.New().String()[:8]
		}
		c.Set(RequestIDKey, id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// RequestLogger logs every HTTP request with method, path, status, and latency.
func RequestLogger() gin.HandlerFunc {
	log := slog.Default().With("component", "HTTP")
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		attrs := []any{
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"ip", c.ClientIP(),
		}

		if reqID, ok := c.Get(RequestIDKey); ok {
			attrs = append(attrs, "request_id", reqID)
		}
		if userID, ok := c.Get("user_id"); ok {
			attrs = append(attrs, "user_id", userID)
		}

		if status >= 500 {
			log.Error("request completed", attrs...)
		} else if status >= 400 {
			log.Warn("request completed", attrs...)
		} else {
			log.Info("request completed", attrs...)
		}
	}
}
