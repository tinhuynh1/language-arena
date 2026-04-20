package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	count    int
	lastSeen time.Time
}

type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(rl.window)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.window {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.mu.Lock()
		v, exists := rl.visitors[ip]
		if !exists || time.Since(v.lastSeen) > rl.window {
			rl.visitors[ip] = &visitor{count: 1, lastSeen: time.Now()}
			rl.mu.Unlock()
			rl.setHeaders(c, 1)
			c.Next()
			return
		}

		v.count++
		v.lastSeen = time.Now()
		count := v.count
		rl.mu.Unlock()

		rl.setHeaders(c, count)

		if count > rl.limit {
			c.JSON(http.StatusTooManyRequests, gin.H{"success": false, "error": "rate limit exceeded"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// setHeaders adds RFC 6585 rate limit headers to every response.
func (rl *RateLimiter) setHeaders(c *gin.Context, currentCount int) {
	remaining := rl.limit - currentCount
	if remaining < 0 {
		remaining = 0
	}
	c.Header("X-RateLimit-Limit", strconv.Itoa(rl.limit))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(rl.window).Unix(), 10))
}
