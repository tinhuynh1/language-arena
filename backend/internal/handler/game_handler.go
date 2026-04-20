package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/michael/language-arena/backend/internal/middleware"
	"github.com/michael/language-arena/backend/internal/repository"
	"github.com/michael/language-arena/backend/internal/service"
	"github.com/michael/language-arena/backend/internal/ws"
	"github.com/michael/language-arena/backend/pkg/response"
)

type GameHandler struct {
	hub      *ws.Hub
	userRepo *repository.UserRepository
	gameRepo *repository.GameRepository
	authSvc  *service.AuthService
	upgrader websocket.Upgrader
	log      *slog.Logger
}

func NewGameHandler(hub *ws.Hub, userRepo *repository.UserRepository, gameRepo *repository.GameRepository, authSvc *service.AuthService, allowedWSOrigins []string) *GameHandler {
	originSet := make(map[string]struct{}, len(allowedWSOrigins))
	for _, o := range allowedWSOrigins {
		originSet[o] = struct{}{}
	}

	return &GameHandler{
		hub:      hub,
		userRepo: userRepo,
		gameRepo: gameRepo,
		authSvc:  authSvc,
		log:      slog.Default().With("component", "HANDLER.Game"),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true
				}
				_, ok := originSet[origin]
				return ok
			},
		},
	}
}

func (h *GameHandler) HandleWebSocket(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		token = c.GetHeader("Authorization")
		if len(token) > 7 {
			token = token[7:]
		}
	}

	userID, err := h.authSvc.ValidateToken(token)
	if err != nil {
		h.log.Warn("ws: invalid token", "err", err, "ip", c.ClientIP(), "request_id", middleware.RequestIDFromContext(c.Request.Context()))
		response.Unauthorized(c, "invalid token")
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("ws: user not found", "user_id", userID, "err", err, "request_id", middleware.RequestIDFromContext(c.Request.Context()))
		response.NotFound(c, "user not found")
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.log.Error("ws: upgrade failed", "user_id", userID, "err", err, "origin", c.Request.Header.Get("Origin"))
		return
	}

	h.log.Info("ws: client connected", "user_id", userID, "username", user.Username, "ip", c.ClientIP())

	client := ws.NewClient(h.hub, conn, user.ID, user.Username)
	h.hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}

func (h *GameHandler) GetOnlineCount(c *gin.Context) {
	response.OK(c, gin.H{"online": h.hub.GetOnlineCount()})
}

func (h *GameHandler) GetGameHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "not authenticated")
		return
	}

	uid := userID.(uuid.UUID)
	games, err := h.gameRepo.FindByUserID(c.Request.Context(), uid, 20)
	if err != nil {
		h.log.Error("get game history failed", "user_id", uid, "err", err, "request_id", middleware.RequestIDFromContext(c.Request.Context()))
		response.InternalError(c, "failed to fetch game history")
		return
	}

	response.OK(c, gin.H{"games": games})
}
