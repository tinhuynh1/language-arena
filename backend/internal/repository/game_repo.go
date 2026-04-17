package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/michael/language-arena/backend/internal/model"
)

type GameRepository struct {
	db *sql.DB
}

func NewGameRepository(db *sql.DB) *GameRepository {
	return &GameRepository{db: db}
}

func (r *GameRepository) Create(ctx context.Context, session *model.GameSession) error {
	session.ID = uuid.New()
	query := `INSERT INTO game_sessions (id, mode, language, player1_id, player2_id, rounds) 
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at`
	return r.db.QueryRowContext(ctx, query,
		session.ID, session.Mode, session.Language,
		session.Player1ID, session.Player2ID, session.Rounds,
	).Scan(&session.CreatedAt)
}

func (r *GameRepository) Finish(ctx context.Context, sessionID uuid.UUID, p1Score, p2Score, avgReaction int, winnerID *uuid.UUID) error {
	query := `UPDATE game_sessions SET 
	          player1_score = $2, player2_score = $3, 
	          avg_reaction_ms = $4, winner_id = $5, 
	          finished_at = NOW() 
	          WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, sessionID, p1Score, p2Score, avgReaction, winnerID)
	return err
}

func (r *GameRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]model.GameSession, error) {
	query := `SELECT id, mode, language, player1_id, player2_id, player1_score, player2_score, 
	          winner_id, rounds, avg_reaction_ms, created_at, finished_at 
	          FROM game_sessions 
	          WHERE player1_id = $1 OR player2_id = $1 
	          ORDER BY created_at DESC LIMIT $2`
	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []model.GameSession
	for rows.Next() {
		var s model.GameSession
		if err := rows.Scan(
			&s.ID, &s.Mode, &s.Language, &s.Player1ID, &s.Player2ID,
			&s.Player1Score, &s.Player2Score, &s.WinnerID, &s.Rounds,
			&s.AvgReactionMs, &s.CreatedAt, &s.FinishedAt,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}
