package i18n

import (
	"github.com/gin-gonic/gin"
)

// LocaleKey is the gin context key where locale is stored.
const LocaleKey = "locale"

// T translates an error code to the given locale.
// Falls back to English if locale or code not found.
func T(locale, code string) string {
	if msgs, ok := messages[code]; ok {
		if msg, ok := msgs[locale]; ok {
			return msg
		}
		if msg, ok := msgs["en"]; ok {
			return msg
		}
	}
	return code
}

// GetLocale retrieves the locale from gin context, defaults to "en".
func GetLocale(c *gin.Context) string {
	if locale, exists := c.Get(LocaleKey); exists {
		if l, ok := locale.(string); ok {
			return l
		}
	}
	return "en"
}

var messages = map[string]map[string]string{
	// Auth middleware
	"err.auth.missing_header": {
		"en": "missing authorization header",
		"vi": "thiếu header xác thực",
	},
	"err.auth.invalid_format": {
		"en": "invalid authorization format",
		"vi": "định dạng xác thực không hợp lệ",
	},
	"err.auth.invalid_token": {
		"en": "invalid or expired token",
		"vi": "token không hợp lệ hoặc đã hết hạn",
	},
	"err.auth.not_authenticated": {
		"en": "not authenticated",
		"vi": "chưa xác thực",
	},

	// Auth handler
	"err.auth.invalid_request": {
		"en": "invalid request",
		"vi": "yêu cầu không hợp lệ",
	},
	"err.auth.email_exists": {
		"en": "email already registered",
		"vi": "email đã được đăng ký",
	},
	"err.auth.username_exists": {
		"en": "username already taken",
		"vi": "tên người dùng đã tồn tại",
	},
	"err.auth.registration_failed": {
		"en": "registration failed",
		"vi": "đăng ký thất bại",
	},
	"err.auth.invalid_credentials": {
		"en": "invalid email or password",
		"vi": "email hoặc mật khẩu không đúng",
	},
	"err.auth.login_failed": {
		"en": "login failed",
		"vi": "đăng nhập thất bại",
	},

	// Game handler
	"err.game.invalid_token": {
		"en": "invalid token",
		"vi": "token không hợp lệ",
	},
	"err.game.user_not_found": {
		"en": "user not found",
		"vi": "không tìm thấy người dùng",
	},
	"err.game.history_failed": {
		"en": "failed to fetch game history",
		"vi": "không thể tải lịch sử trận đấu",
	},

	// Leaderboard handler
	"err.leaderboard.fetch_failed": {
		"en": "failed to fetch leaderboard",
		"vi": "không thể tải bảng xếp hạng",
	},
	"err.leaderboard.stats_failed": {
		"en": "failed to fetch stats",
		"vi": "không thể tải thống kê",
	},

	// Vocab handler
	"err.vocab.invalid_query": {
		"en": "invalid query: lang must be 'en' or 'zh'",
		"vi": "truy vấn không hợp lệ: ngôn ngữ phải là 'en' hoặc 'zh'",
	},
	"err.vocab.fetch_failed": {
		"en": "failed to fetch vocabularies",
		"vi": "không thể tải từ vựng",
	},
}
