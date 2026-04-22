package i18n

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ── T() Tests ──────────────────────────────────────────

func TestT_EnglishTranslation(t *testing.T) {
	msg := T("en", "err.auth.missing_header")
	assert.Equal(t, "missing authorization header", msg)
}

func TestT_VietnameseTranslation(t *testing.T) {
	msg := T("vi", "err.auth.missing_header")
	assert.Equal(t, "thiếu header xác thực", msg)
}

func TestT_UnknownLocale_FallsBackToEnglish(t *testing.T) {
	msg := T("fr", "err.auth.missing_header")
	assert.Equal(t, "missing authorization header", msg)
}

func TestT_UnknownCode_ReturnsCode(t *testing.T) {
	msg := T("en", "err.unknown.code")
	assert.Equal(t, "err.unknown.code", msg)
}

func TestT_AllCodesHaveEnglish(t *testing.T) {
	for code, locales := range messages {
		enMsg, ok := locales["en"]
		assert.True(t, ok, "code %q missing English translation", code)
		assert.NotEmpty(t, enMsg, "code %q has empty English translation", code)
	}
}

func TestT_AllCodesHaveVietnamese(t *testing.T) {
	for code, locales := range messages {
		viMsg, ok := locales["vi"]
		assert.True(t, ok, "code %q missing Vietnamese translation", code)
		assert.NotEmpty(t, viMsg, "code %q has empty Vietnamese translation", code)
	}
}

// ── GetLocale() Tests ──────────────────────────────────

func TestGetLocale_WithLocaleSet(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set(LocaleKey, "vi")

	locale := GetLocale(c)
	assert.Equal(t, "vi", locale)
}

func TestGetLocale_NoLocale_DefaultsToEn(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	locale := GetLocale(c)
	assert.Equal(t, "en", locale)
}

func TestGetLocale_WrongType_DefaultsToEn(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set(LocaleKey, 42) // wrong type

	locale := GetLocale(c)
	assert.Equal(t, "en", locale)
}
