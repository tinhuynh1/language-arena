package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/michael/language-arena/backend/pkg/i18n"
)

// LocaleMiddleware extracts the locale from Accept-Language header.
// Supports "en" and "vi", defaults to "en".
func LocaleMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := "en"
		accept := c.GetHeader("Accept-Language")
		if accept != "" {
			primary := strings.SplitN(accept, ",", 2)[0]
			primary = strings.TrimSpace(strings.SplitN(primary, ";", 2)[0])
			if strings.HasPrefix(primary, "vi") {
				lang = "vi"
			}
		}
		c.Set(i18n.LocaleKey, lang)
		c.Next()
	}
}
