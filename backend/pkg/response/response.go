package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/michael/language-arena/backend/pkg/i18n"
)

type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	ErrorCode string      `json:"error_code,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
}

type Meta struct {
	Total  int `json:"total,omitempty"`
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{Success: true, Data: data})
}

func WithMeta(c *gin.Context, data interface{}, meta *Meta) {
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: data, Meta: meta})
}

func BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: msg})
}

func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, APIResponse{Success: false, Error: msg})
}

func NotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, APIResponse{Success: false, Error: msg})
}

func InternalError(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: msg})
}

// Localized error responses — translate error code based on Accept-Language

func BadRequestI18n(c *gin.Context, code string) {
	locale := i18n.GetLocale(c)
	msg := i18n.T(locale, code)
	c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: msg, ErrorCode: code})
}

func UnauthorizedI18n(c *gin.Context, code string) {
	locale := i18n.GetLocale(c)
	msg := i18n.T(locale, code)
	c.JSON(http.StatusUnauthorized, APIResponse{Success: false, Error: msg, ErrorCode: code})
}

func NotFoundI18n(c *gin.Context, code string) {
	locale := i18n.GetLocale(c)
	msg := i18n.T(locale, code)
	c.JSON(http.StatusNotFound, APIResponse{Success: false, Error: msg, ErrorCode: code})
}

func InternalErrorI18n(c *gin.Context, code string) {
	locale := i18n.GetLocale(c)
	msg := i18n.T(locale, code)
	c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: msg, ErrorCode: code})
}
