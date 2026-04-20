package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDKey = "request_id"

type contextKey string

const requestIDCtxKey contextKey = "request_id"

// RequestIDFromContext extracts request_id from context for downstream logging.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDCtxKey).(string); ok {
		return id
	}
	return ""
}

// RequestID injects a unique X-Request-ID into each request and propagates it into context.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.New().String()[:8]
		}
		c.Set(RequestIDKey, id)
		c.Header("X-Request-ID", id)

		// Propagate into context.Context for downstream use (repos, services)
		ctx := context.WithValue(c.Request.Context(), requestIDCtxKey, id)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// RequestLogger logs every HTTP request with method, path, status, latency, and body size.
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
			"response_bytes", c.Writer.Size(),
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
