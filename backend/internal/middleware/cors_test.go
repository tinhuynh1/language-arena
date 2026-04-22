package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORSMiddleware([]string{"http://allowed.com"}))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// Test GET request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://allowed.com")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://allowed.com", w.Header().Get("Access-Control-Allow-Origin"))

	// Test OPTIONS preflight
	wOptions := httptest.NewRecorder()
	reqOptions, _ := http.NewRequest("OPTIONS", "/test", nil)
	reqOptions.Header.Set("Origin", "http://allowed.com")
	router.ServeHTTP(wOptions, reqOptions)

	assert.Equal(t, http.StatusNoContent, wOptions.Code)
	assert.Equal(t, "http://allowed.com", wOptions.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, wOptions.Header().Get("Access-Control-Allow-Methods"), "POST")
}
