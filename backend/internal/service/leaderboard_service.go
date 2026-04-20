package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/michael/language-arena/backend/internal/model"
)

// UserStatsReader abstracts user data reads for leaderboard.
type UserStatsReader interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetLeaderboard(ctx context.Context, limit int) ([]model.LeaderboardEntry, error)
}

// GameReader abstracts game session reads.
type GameReader interface {
	FindByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]model.GameSession, error)
}

type LeaderboardService struct {
	userReader UserStatsReader
	gameReader GameReader
}

func NewLeaderboardService(userReader UserStatsReader, gameReader GameReader) *LeaderboardService {
	return &LeaderboardService{userReader: userReader, gameReader: gameReader}
}

func (s *LeaderboardService) GetTopPlayers(ctx context.Context, limit int) ([]model.LeaderboardEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.userReader.GetLeaderboard(ctx, limit)
}

func (s *LeaderboardService) GetPlayerStats(ctx context.Context, userID uuid.UUID) (*model.User, []model.GameSession, error) {
	user, err := s.userReader.FindByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	games, err := s.gameReader.FindByUserID(ctx, userID, 10)
	if err != nil {
		return user, nil, err
	}

	return user, games, nil
}
