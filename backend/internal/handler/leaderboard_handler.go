package handler

import (
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/michael/language-arena/backend/internal/middleware"
	"github.com/michael/language-arena/backend/internal/service"
	"github.com/michael/language-arena/backend/pkg/response"
)

type LeaderboardHandler struct {
	leaderboardService *service.LeaderboardService
	log                *slog.Logger
}

func NewLeaderboardHandler(leaderboardService *service.LeaderboardService) *LeaderboardHandler {
	return &LeaderboardHandler{
		leaderboardService: leaderboardService,
		log:                slog.Default().With("component", "HANDLER.Leaderboard"),
	}
}

func (h *LeaderboardHandler) GetLeaderboard(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	entries, err := h.leaderboardService.GetTopPlayers(c.Request.Context(), limit)
	if err != nil {
		h.log.Error("get leaderboard failed", "limit", limit, "err", err, "request_id", middleware.RequestIDFromContext(c.Request.Context()))
		response.InternalError(c, "failed to fetch leaderboard")
		return
	}

	response.OK(c, entries)
}

func (h *LeaderboardHandler) GetMyStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "not authenticated")
		return
	}

	uid := userID.(uuid.UUID)
	user, games, err := h.leaderboardService.GetPlayerStats(c.Request.Context(), uid)
	if err != nil {
		h.log.Error("get player stats failed", "user_id", uid, "err", err, "request_id", middleware.RequestIDFromContext(c.Request.Context()))
		response.InternalError(c, "failed to fetch stats")
		return
	}

	response.OK(c, gin.H{
		"user":         user,
		"recent_games": games,
	})
}
