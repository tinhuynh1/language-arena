package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/michael/language-arena/backend/internal/model"
	"github.com/michael/language-arena/backend/internal/service"
	"github.com/michael/language-arena/backend/pkg/response"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	result, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		switch err {
		case service.ErrUserExists:
			response.BadRequest(c, err.Error())
		case service.ErrUsernameExists:
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, "registration failed")
		}
		return
	}

	response.Created(c, result)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	result, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": err.Error()})
			return
		}
		response.InternalError(c, "login failed")
		return
	}

	response.OK(c, result)
}
