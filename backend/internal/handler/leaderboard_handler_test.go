package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/michael/language-arena/backend/internal/model"
	"github.com/michael/language-arena/backend/internal/service"
	"github.com/michael/language-arena/backend/pkg/response"
	"github.com/stretchr/testify/assert"
)

// ── Mocks ──────────────────────────────────────────────

type mockLBUserStatsReader struct {
	findByIDFn       func(ctx context.Context, id uuid.UUID) (*model.User, error)
	getLeaderboardFn func(ctx context.Context, limit int, offset int) ([]model.LeaderboardEntry, int, error)
}

func (m *mockLBUserStatsReader) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockLBUserStatsReader) GetLeaderboard(ctx context.Context, limit int, offset int) ([]model.LeaderboardEntry, int, error) {
	return m.getLeaderboardFn(ctx, limit, offset)
}

type mockLBGameReader struct {
	findByUserIDFn func(ctx context.Context, userID uuid.UUID, limit int) ([]model.GameSession, error)
}

func (m *mockLBGameReader) FindByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]model.GameSession, error) {
	return m.findByUserIDFn(ctx, userID, limit)
}

func parseLBResponse(t *testing.T, w *httptest.ResponseRecorder) response.APIResponse {
	var resp response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	return resp
}

// ── GetLeaderboard Tests ───────────────────────────────

func TestGetLeaderboard_Success(t *testing.T) {
	entries := []model.LeaderboardEntry{
		{Rank: 1, Username: "top1", AvgReactionMs: 200},
	}
	userReader := &mockLBUserStatsReader{
		getLeaderboardFn: func(_ context.Context, _ int, _ int) ([]model.LeaderboardEntry, int, error) {
			return entries, 1, nil
		},
	}

	svc := service.NewLeaderboardService(userReader, nil)
	h := NewLeaderboardHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/leaderboard?limit=10&page=1", nil)

	h.GetLeaderboard(c)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseLBResponse(t, w)
	assert.True(t, resp.Success)
}

func TestGetLeaderboard_InternalError(t *testing.T) {
	userReader := &mockLBUserStatsReader{
		getLeaderboardFn: func(_ context.Context, _ int, _ int) ([]model.LeaderboardEntry, int, error) {
			return nil, 0, errors.New("db down")
		},
	}

	svc := service.NewLeaderboardService(userReader, nil)
	h := NewLeaderboardHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/leaderboard", nil)

	h.GetLeaderboard(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ── GetMyStats Tests ───────────────────────────────────

func TestGetMyStats_Success(t *testing.T) {
	userID := uuid.New()
	user := &model.User{ID: userID, Username: "player"}
	games := []model.GameSession{{ID: uuid.New(), Mode: model.ModeSolo}}

	userReader := &mockLBUserStatsReader{
		findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.User, error) { return user, nil },
	}
	gameReader := &mockLBGameReader{
		findByUserIDFn: func(_ context.Context, _ uuid.UUID, _ int) ([]model.GameSession, error) { return games, nil },
	}

	svc := service.NewLeaderboardService(userReader, gameReader)
	h := NewLeaderboardHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/stats/me", nil)
	c.Set("user_id", userID)

	h.GetMyStats(c)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseLBResponse(t, w)
	assert.True(t, resp.Success)
}

func TestGetMyStats_Unauthenticated(t *testing.T) {
	svc := service.NewLeaderboardService(nil, nil)
	h := NewLeaderboardHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/stats/me", nil)
	// don't set user_id

	h.GetMyStats(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
