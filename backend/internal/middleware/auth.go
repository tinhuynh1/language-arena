package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/michael/language-arena/backend/internal/service"
	"github.com/michael/language-arena/backend/pkg/response"
)

func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.UnauthorizedI18n(c, "err.auth.missing_header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.UnauthorizedI18n(c, "err.auth.invalid_format")
			c.Abort()
			return
		}

		userID, err := authService.ValidateToken(parts[1])
		if err != nil {
			response.UnauthorizedI18n(c, "err.auth.invalid_token")
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
