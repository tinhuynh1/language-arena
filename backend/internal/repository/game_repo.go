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
	query := `INSERT INTO game_sessions (id, mode, language, rounds) 
	          VALUES ($1, $2, $3, $4) RETURNING created_at`
	return r.db.QueryRowContext(ctx, query,
		session.ID, session.Mode, session.Language, session.Rounds,
	).Scan(&session.CreatedAt)
}

func (r *GameRepository) Finish(ctx context.Context, sessionID uuid.UUID, avgReaction int, winnerID *uuid.UUID) error {
	query := `UPDATE game_sessions SET 
	          avg_reaction_ms = $2, winner_id = $3, 
	          finished_at = NOW() 
	          WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, sessionID, avgReaction, winnerID)
	return err
}

func (r *GameRepository) CreatePlayerResult(ctx context.Context, p *model.GameSessionPlayer) error {
	p.ID = uuid.New()
	query := `INSERT INTO game_session_players (id, session_id, user_id, score, avg_reaction_ms, best_reaction_ms, rank)
	          VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.SessionID, p.UserID, p.Score, p.AvgReactionMs, p.BestReactionMs, p.Rank,
	)
	return err
}

func (r *GameRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]model.GameSession, error) {
	query := `SELECT DISTINCT gs.id, gs.mode, gs.language, gs.winner_id, gs.rounds, 
	          gs.avg_reaction_ms, gs.created_at, gs.finished_at 
	          FROM game_sessions gs
	          INNER JOIN game_session_players gsp ON gs.id = gsp.session_id
	          WHERE gsp.user_id = $1
	          ORDER BY gs.created_at DESC LIMIT $2`
	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []model.GameSession
	for rows.Next() {
		var s model.GameSession
		if err := rows.Scan(
			&s.ID, &s.Mode, &s.Language, &s.WinnerID, &s.Rounds,
			&s.AvgReactionMs, &s.CreatedAt, &s.FinishedAt,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}
