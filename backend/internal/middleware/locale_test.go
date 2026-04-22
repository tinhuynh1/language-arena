package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLocaleMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(LocaleMiddleware())
	router.GET("/test", func(c *gin.Context) {
		locale, _ := c.Get("locale")
		c.String(http.StatusOK, locale.(string))
	})

	tests := []struct {
		header   string
		expected string
	}{
		{"vi,en;q=0.9", "vi"},
		{"en-US,en;q=0.9", "en"},
		{"fr-CH, fr;q=0.9, en;q=0.8", "en"}, // Unknown falls back to en
		{"", "en"}, // Empty falls back to en
	}

	for _, tt := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		if tt.header != "" {
			req.Header.Set("Accept-Language", tt.header)
		}
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, tt.expected, w.Body.String())
	}
}
