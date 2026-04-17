package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/michael/language-arena/backend/internal/model"
	"github.com/michael/language-arena/backend/internal/repository"
)

type LeaderboardService struct {
	userRepo *repository.UserRepository
	gameRepo *repository.GameRepository
}

func NewLeaderboardService(userRepo *repository.UserRepository, gameRepo *repository.GameRepository) *LeaderboardService {
	return &LeaderboardService{userRepo: userRepo, gameRepo: gameRepo}
}

func (s *LeaderboardService) GetTopPlayers(ctx context.Context, limit int) ([]model.LeaderboardEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.userRepo.GetLeaderboard(ctx, limit)
}

func (s *LeaderboardService) GetPlayerStats(ctx context.Context, userID uuid.UUID) (*model.User, []model.GameSession, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	games, err := s.gameRepo.FindByUserID(ctx, userID, 10)
	if err != nil {
		return user, nil, err
	}

	return user, games, nil
}
