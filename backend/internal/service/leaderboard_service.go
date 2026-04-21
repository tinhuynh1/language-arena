package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/michael/language-arena/backend/internal/model"
)

// UserStatsReader abstracts user data reads for leaderboard.
type UserStatsReader interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetLeaderboard(ctx context.Context, limit int, offset int) ([]model.LeaderboardEntry, int, error)
}

// GameReader abstracts game session reads.
type GameReader interface {
	FindByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]model.GameSession, error)
}

type LeaderboardService struct {
	userReader UserStatsReader
	gameReader GameReader
	log        *slog.Logger
}

func NewLeaderboardService(userReader UserStatsReader, gameReader GameReader) *LeaderboardService {
	return &LeaderboardService{
		userReader: userReader,
		gameReader: gameReader,
		log:        slog.Default().With("component", "SVC.Leaderboard"),
	}
}

func (s *LeaderboardService) GetTopPlayers(ctx context.Context, limit int, page int) ([]model.LeaderboardEntry, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit
	entries, total, err := s.userReader.GetLeaderboard(ctx, limit, offset)
	if err != nil {
		s.log.Error("get top players failed", "op", "GetTopPlayers", "limit", limit, "page", page, "err", err)
		return nil, 0, err
	}

	s.log.Debug("leaderboard fetched", "op", "GetTopPlayers", "limit", limit, "page", page, "count", len(entries), "total", total)
	return entries, total, nil
}

func (s *LeaderboardService) GetPlayerStats(ctx context.Context, userID uuid.UUID) (*model.User, []model.GameSession, error) {
	user, err := s.userReader.FindByID(ctx, userID)
	if err != nil {
		s.log.Error("get player stats failed: user not found", "op", "GetPlayerStats", "user_id", userID, "err", err)
		return nil, nil, err
	}

	games, err := s.gameReader.FindByUserID(ctx, userID, 10)
	if err != nil {
		s.log.Error("get player stats failed: games query error", "op", "GetPlayerStats", "user_id", userID, "err", err)
		return user, nil, err
	}

	s.log.Debug("player stats fetched", "op", "GetPlayerStats", "user_id", userID, "games_count", len(games))
	return user, games, nil
}
