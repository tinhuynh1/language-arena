package repository

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/michael/language-arena/backend/internal/model"
)

type GameRepository struct {
	db  *sql.DB
	log *slog.Logger
}

func NewGameRepository(db *sql.DB) *GameRepository {
	return &GameRepository{
		db:  db,
		log: slog.Default().With("component", "REPO.Game"),
	}
}

func (r *GameRepository) Create(ctx context.Context, session *model.GameSession) error {
	start := time.Now()
	session.ID = uuid.New()
	query := `INSERT INTO game_sessions (id, mode, language, rounds) 
	          VALUES ($1, $2, $3, $4) RETURNING created_at`
	err := r.db.QueryRowContext(ctx, query,
		session.ID, session.Mode, session.Language, session.Rounds,
	).Scan(&session.CreatedAt)

	duration := time.Since(start)
	if err != nil {
		r.log.Error("create game session failed", "op", "Create", "mode", session.Mode, "language", session.Language, "err", err, "duration_ms", duration.Milliseconds())
		return err
	}
	if duration > slowQueryThreshold {
		r.log.Warn("slow query", "op", "Create", "session_id", session.ID, "duration_ms", duration.Milliseconds())
	}
	return nil
}

func (r *GameRepository) Finish(ctx context.Context, sessionID uuid.UUID, avgReaction int, winnerID *uuid.UUID) error {
	start := time.Now()
	query := `UPDATE game_sessions SET 
	          avg_reaction_ms = $2, winner_id = $3, 
	          finished_at = NOW() 
	          WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, sessionID, avgReaction, winnerID)

	duration := time.Since(start)
	if err != nil {
		r.log.Error("finish game session failed", "op", "Finish", "session_id", sessionID, "err", err, "duration_ms", duration.Milliseconds())
		return err
	}
	if duration > slowQueryThreshold {
		r.log.Warn("slow query", "op", "Finish", "session_id", sessionID, "duration_ms", duration.Milliseconds())
	}
	return nil
}

func (r *GameRepository) CreatePlayerResult(ctx context.Context, p *model.GameSessionPlayer) error {
	start := time.Now()
	p.ID = uuid.New()
	query := `INSERT INTO game_session_players (id, session_id, user_id, score, avg_reaction_ms, best_reaction_ms, rank)
	          VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.SessionID, p.UserID, p.Score, p.AvgReactionMs, p.BestReactionMs, p.Rank,
	)

	duration := time.Since(start)
	if err != nil {
		r.log.Error("create player result failed", "op", "CreatePlayerResult", "session_id", p.SessionID, "user_id", p.UserID, "err", err, "duration_ms", duration.Milliseconds())
		return err
	}
	if duration > slowQueryThreshold {
		r.log.Warn("slow query", "op", "CreatePlayerResult", "session_id", p.SessionID, "duration_ms", duration.Milliseconds())
	}
	return nil
}

func (r *GameRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]model.GameSession, error) {
	start := time.Now()
	query := `SELECT DISTINCT gs.id, gs.mode, gs.language, gs.winner_id, gs.rounds, 
	          gs.avg_reaction_ms, gs.created_at, gs.finished_at 
	          FROM game_sessions gs
	          INNER JOIN game_session_players gsp ON gs.id = gsp.session_id
	          WHERE gsp.user_id = $1
	          ORDER BY gs.created_at DESC LIMIT $2`
	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		duration := time.Since(start)
		r.log.Error("find games by user failed", "op", "FindByUserID", "user_id", userID, "limit", limit, "err", err, "duration_ms", duration.Milliseconds())
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
			r.log.Error("scan game session row failed", "op", "FindByUserID", "user_id", userID, "err", err)
			return nil, err
		}
		sessions = append(sessions, s)
	}

	duration := time.Since(start)
	if duration > slowQueryThreshold {
		r.log.Warn("slow query", "op", "FindByUserID", "user_id", userID, "duration_ms", duration.Milliseconds())
	}
	return sessions, rows.Err()
}
