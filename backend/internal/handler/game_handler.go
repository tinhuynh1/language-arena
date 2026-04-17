package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/michael/language-arena/backend/internal/repository"
	"github.com/michael/language-arena/backend/internal/service"
	"github.com/michael/language-arena/backend/internal/ws"
	"github.com/michael/language-arena/backend/pkg/response"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type GameHandler struct {
	hub      *ws.Hub
	userRepo *repository.UserRepository
	authSvc  *service.AuthService
}

func NewGameHandler(hub *ws.Hub, userRepo *repository.UserRepository, authSvc *service.AuthService) *GameHandler {
	return &GameHandler{hub: hub, userRepo: userRepo, authSvc: authSvc}
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

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
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

	gameRepo := repository.NewGameRepository(nil)
	_ = gameRepo
	_ = userID.(uuid.UUID)

	response.OK(c, gin.H{"games": []interface{}{}})
}
