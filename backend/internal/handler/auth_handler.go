package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/michael/language-arena/backend/internal/middleware"
	"github.com/michael/language-arena/backend/internal/model"
	"github.com/michael/language-arena/backend/internal/service"
	"github.com/michael/language-arena/backend/pkg/response"
)

type AuthHandler struct {
	authService *service.AuthService
	log         *slog.Logger
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		log:         slog.Default().With("component", "HANDLER.Auth"),
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("register: invalid request body", "err", err, "request_id", middleware.RequestIDFromContext(c.Request.Context()))
		response.BadRequestI18n(c, "err.auth.invalid_request")
		return
	}

	result, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		reqID := middleware.RequestIDFromContext(c.Request.Context())
		switch err {
		case service.ErrUserExists:
			h.log.Info("register: email already exists", "email", req.Email, "request_id", reqID)
			response.BadRequestI18n(c, "err.auth.email_exists")
		case service.ErrUsernameExists:
			h.log.Info("register: username already taken", "username", req.Username, "request_id", reqID)
			response.BadRequestI18n(c, "err.auth.username_exists")
		default:
			h.log.Error("register: internal error", "email", req.Email, "err", err, "request_id", reqID)
			response.InternalErrorI18n(c, "err.auth.registration_failed")
		}
		return
	}

	response.Created(c, result)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("login: invalid request body", "err", err, "request_id", middleware.RequestIDFromContext(c.Request.Context()))
		response.BadRequestI18n(c, "err.auth.invalid_request")
		return
	}

	result, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		reqID := middleware.RequestIDFromContext(c.Request.Context())
		if err == service.ErrInvalidCredentials {
			h.log.Info("login: invalid credentials", "email", req.Email, "request_id", reqID)
			response.UnauthorizedI18n(c, "err.auth.invalid_credentials")
			return
		}
		h.log.Error("login: internal error", "email", req.Email, "err", err, "request_id", reqID)
		response.InternalErrorI18n(c, "err.auth.login_failed")
		return
	}

	response.OK(c, result)
}
