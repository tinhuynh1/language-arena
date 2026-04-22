package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/michael/language-arena/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Mocks ──────────────────────────────────────────────

type mockUserStatsReader struct {
	findByIDFn       func(ctx context.Context, id uuid.UUID) (*model.User, error)
	getLeaderboardFn func(ctx context.Context, limit int, offset int) ([]model.LeaderboardEntry, int, error)
}

func (m *mockUserStatsReader) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockUserStatsReader) GetLeaderboard(ctx context.Context, limit int, offset int) ([]model.LeaderboardEntry, int, error) {
	return m.getLeaderboardFn(ctx, limit, offset)
}

type mockGameReader struct {
	findByUserIDFn func(ctx context.Context, userID uuid.UUID, limit int) ([]model.GameSession, error)
}

func (m *mockGameReader) FindByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]model.GameSession, error) {
	return m.findByUserIDFn(ctx, userID, limit)
}

// ── GetTopPlayers Tests ────────────────────────────────

func TestGetTopPlayers_Success(t *testing.T) {
	entries := []model.LeaderboardEntry{
		{Rank: 1, Username: "player1", AvgReactionMs: 300},
		{Rank: 2, Username: "player2", AvgReactionMs: 450},
	}
	userReader := &mockUserStatsReader{
		getLeaderboardFn: func(_ context.Context, limit int, offset int) ([]model.LeaderboardEntry, int, error) {
			assert.Equal(t, 10, limit)
			assert.Equal(t, 0, offset)
			return entries, 50, nil
		},
	}

	svc := NewLeaderboardService(userReader, nil)
	result, total, err := svc.GetTopPlayers(context.Background(), 10, 1)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 50, total)
	assert.Equal(t, "player1", result[0].Username)
}

func TestGetTopPlayers_LimitClamping(t *testing.T) {
	tests := []struct {
		name          string
		inputLimit    int
		expectedLimit int
	}{
		{"zero limit defaults to 10", 0, 10},
		{"negative limit defaults to 10", -5, 10},
		{"over 100 defaults to 10", 101, 10},
		{"valid limit passes through", 25, 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userReader := &mockUserStatsReader{
				getLeaderboardFn: func(_ context.Context, limit int, _ int) ([]model.LeaderboardEntry, int, error) {
					assert.Equal(t, tt.expectedLimit, limit)
					return nil, 0, nil
				},
			}

			svc := NewLeaderboardService(userReader, nil)
			_, _, err := svc.GetTopPlayers(context.Background(), tt.inputLimit, 1)
			require.NoError(t, err)
		})
	}
}

func TestGetTopPlayers_PageNormalization(t *testing.T) {
	tests := []struct {
		name           string
		inputPage      int
		expectedOffset int
	}{
		{"page 1 → offset 0", 1, 0},
		{"page 2 → offset 10", 2, 10},
		{"page 0 → normalized to 1 → offset 0", 0, 0},
		{"negative page → normalized to 1 → offset 0", -3, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userReader := &mockUserStatsReader{
				getLeaderboardFn: func(_ context.Context, _ int, offset int) ([]model.LeaderboardEntry, int, error) {
					assert.Equal(t, tt.expectedOffset, offset)
					return nil, 0, nil
				},
			}

			svc := NewLeaderboardService(userReader, nil)
			_, _, err := svc.GetTopPlayers(context.Background(), 10, tt.inputPage)
			require.NoError(t, err)
		})
	}
}

func TestGetTopPlayers_RepoError(t *testing.T) {
	dbErr := errors.New("database connection lost")
	userReader := &mockUserStatsReader{
		getLeaderboardFn: func(_ context.Context, _ int, _ int) ([]model.LeaderboardEntry, int, error) {
			return nil, 0, dbErr
		},
	}

	svc := NewLeaderboardService(userReader, nil)
	result, total, err := svc.GetTopPlayers(context.Background(), 10, 1)

	assert.Nil(t, result)
	assert.Equal(t, 0, total)
	assert.ErrorIs(t, err, dbErr)
}

// ── GetPlayerStats Tests ───────────────────────────────

func TestGetPlayerStats_Success(t *testing.T) {
	userID := uuid.New()
	user := &model.User{ID: userID, Username: "statsplayer", GamesPlayed: 15}
	games := []model.GameSession{{ID: uuid.New(), Mode: model.ModeSolo}}

	userReader := &mockUserStatsReader{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*model.User, error) {
			assert.Equal(t, userID, id)
			return user, nil
		},
	}
	gameReader := &mockGameReader{
		findByUserIDFn: func(_ context.Context, uid uuid.UUID, limit int) ([]model.GameSession, error) {
			assert.Equal(t, userID, uid)
			assert.Equal(t, 10, limit)
			return games, nil
		},
	}

	svc := NewLeaderboardService(userReader, gameReader)
	u, g, err := svc.GetPlayerStats(context.Background(), userID)

	require.NoError(t, err)
	assert.Equal(t, "statsplayer", u.Username)
	assert.Len(t, g, 1)
}

func TestGetPlayerStats_UserNotFound(t *testing.T) {
	userReader := &mockUserStatsReader{
		findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.User, error) {
			return nil, errors.New("user not found")
		},
	}

	svc := NewLeaderboardService(userReader, nil)
	u, g, err := svc.GetPlayerStats(context.Background(), uuid.New())

	assert.Nil(t, u)
	assert.Nil(t, g)
	assert.Error(t, err)
}

func TestGetPlayerStats_GamesFetchError(t *testing.T) {
	userID := uuid.New()
	user := &model.User{ID: userID, Username: "player"}

	userReader := &mockUserStatsReader{
		findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.User, error) {
			return user, nil
		},
	}
	gameReader := &mockGameReader{
		findByUserIDFn: func(_ context.Context, _ uuid.UUID, _ int) ([]model.GameSession, error) {
			return nil, errors.New("query timeout")
		},
	}

	svc := NewLeaderboardService(userReader, gameReader)
	u, g, err := svc.GetPlayerStats(context.Background(), userID)

	assert.NotNil(t, u) // user is returned even on games error
	assert.Nil(t, g)
	assert.Error(t, err)
}
