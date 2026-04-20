package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true // Allow non-browser clients (curl, Postman)
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
		response.Unauthorized(c, "invalid token")
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("ws upgrade error", "component", "WS", "err", err, "user_id", userID)
		return
	}

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

	games, err := h.gameRepo.FindByUserID(c.Request.Context(), userID.(uuid.UUID), 20)
	if err != nil {
		response.InternalError(c, "failed to fetch game history")
		return
	}

	response.OK(c, gin.H{"games": games})
}
