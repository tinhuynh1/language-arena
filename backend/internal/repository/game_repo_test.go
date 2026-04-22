package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/michael/language-arena/backend/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestGameRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewGameRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		session := &model.GameSession{
			Mode:     "duel",
			Language: "en",
			Rounds:   5,
		}

		mock.ExpectQuery("INSERT INTO game_sessions").
			WithArgs(sqlmock.AnyArg(), session.Mode, session.Language, session.Rounds).
			WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(time.Now()))

		err := repo.Create(ctx, session)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, session.ID)
	})

	t.Run("error", func(t *testing.T) {
		session := &model.GameSession{Mode: "duel"}
		mock.ExpectQuery("INSERT INTO game_sessions").WillReturnError(sql.ErrConnDone)

		err := repo.Create(ctx, session)
		assert.ErrorIs(t, err, sql.ErrConnDone)
	})
}

func TestGameRepository_Finish(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewGameRepository(db)
	ctx := context.Background()
	sessionID := uuid.New()
	winnerID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec("UPDATE game_sessions SET").
			WithArgs(sessionID, 1200, &winnerID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Finish(ctx, sessionID, 1200, &winnerID)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mock.ExpectExec("UPDATE game_sessions SET").WillReturnError(sql.ErrConnDone)

		err := repo.Finish(ctx, sessionID, 1200, &winnerID)
		assert.ErrorIs(t, err, sql.ErrConnDone)
	})
}

func TestGameRepository_CreatePlayerResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewGameRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		p := &model.GameSessionPlayer{
			SessionID:      uuid.New(),
			UserID:         uuid.New(),
			Score:          100,
			AvgReactionMs:  1200,
			BestReactionMs: 800,
			Rank:           1,
		}

		mock.ExpectExec("INSERT INTO game_session_players").
			WithArgs(sqlmock.AnyArg(), p.SessionID, p.UserID, p.Score, p.AvgReactionMs, p.BestReactionMs, p.Rank).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.CreatePlayerResult(ctx, p)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, p.ID)
	})
}

func TestGameRepository_FindByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewGameRepository(db)
	ctx := context.Background()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "mode", "language", "winner_id", "rounds", "avg_reaction_ms", "created_at", "finished_at"}).
			AddRow(uuid.New(), "duel", "en", nil, 5, 1200, time.Now(), time.Now())

		mock.ExpectQuery("SELECT DISTINCT gs.id").WithArgs(userID, 10).WillReturnRows(rows)

		sessions, err := repo.FindByUserID(ctx, userID, 10)
		assert.NoError(t, err)
		assert.Len(t, sessions, 1)
	})
}
