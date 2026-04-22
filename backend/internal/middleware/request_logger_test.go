package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// test context fallback
	assert.Empty(t, RequestIDFromContext(context.Background()))

	router := gin.New()
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		reqID, exists := c.Get(RequestIDKey)
		assert.True(t, exists)
		assert.NotEmpty(t, reqID)

		ctxReqID := RequestIDFromContext(c.Request.Context())
		assert.Equal(t, reqID.(string), ctxReqID)

		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	// Try providing existing header
	req.Header.Set("X-Request-ID", "custom-id")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "custom-id", w.Header().Get("X-Request-ID"))

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.NotEmpty(t, w2.Header().Get("X-Request-ID"))
	assert.NotEqual(t, "custom-id", w2.Header().Get("X-Request-ID"))
}

func TestRequestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestID())
	router.Use(RequestLogger())
	
	router.GET("/200", func(c *gin.Context) {
		c.Set("user_id", "123")
		c.String(http.StatusOK, "ok")
	})
	router.GET("/400", func(c *gin.Context) { c.String(http.StatusBadRequest, "bad") })
	router.GET("/500", func(c *gin.Context) { c.String(http.StatusInternalServerError, "err") })

	tests := []struct {
		path string
		code int
	}{
		{"/200", 200},
		{"/400", 400},
		{"/500", 500},
	}

	for _, tt := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", tt.path, nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, tt.code, w.Code)
	}
}
