package repository

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/michael/language-arena/backend/internal/model"
)

const slowQueryThreshold = 100 * time.Millisecond

type UserRepository struct {
	db  *sql.DB
	log *slog.Logger
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db:  db,
		log: slog.Default().With("component", "REPO.User"),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	start := time.Now()
	query := `INSERT INTO users (id, username, email, password_hash)
	          VALUES ($1, $2, $3, $4) RETURNING created_at`
	user.ID = uuid.New()
	err := r.db.QueryRowContext(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash,
	).Scan(&user.CreatedAt)

	duration := time.Since(start)
	if err != nil {
		r.log.Error("create user failed", "op", "Create", "email", user.Email, "err", err, "duration_ms", duration.Milliseconds(), r.reqIDAttr(ctx))
		return err
	}
	r.warnSlow("Create", duration, ctx)
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	start := time.Now()
	user := &model.User{}
	query := `SELECT id, username, email, password_hash, avg_reaction_ms, total_correct, games_played, best_reaction_ms, created_at
	          FROM users WHERE email = $1`
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.AvgReactionMs, &user.TotalCorrect, &user.GamesPlayed, &user.BestReactionMs, &user.CreatedAt,
	)

	duration := time.Since(start)
	if err != nil {
		if err != sql.ErrNoRows {
			r.log.Error("find user by email failed", "op", "FindByEmail", "email", email, "err", err, "duration_ms", duration.Milliseconds(), r.reqIDAttr(ctx))
		}
		return nil, err
	}
	r.warnSlow("FindByEmail", duration, ctx)
	return user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	start := time.Now()
	user := &model.User{}
	query := `SELECT id, username, email, password_hash, avg_reaction_ms, total_correct, games_played, best_reaction_ms, created_at
	          FROM users WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.AvgReactionMs, &user.TotalCorrect, &user.GamesPlayed, &user.BestReactionMs, &user.CreatedAt,
	)

	duration := time.Since(start)
	if err != nil {
		if err != sql.ErrNoRows {
			r.log.Error("find user by id failed", "op", "FindByID", "user_id", id, "err", err, "duration_ms", duration.Milliseconds(), r.reqIDAttr(ctx))
		}
		return nil, err
	}
	r.warnSlow("FindByID", duration, ctx)
	return user, nil
}

func (r *UserRepository) UpdateStats(ctx context.Context, userID uuid.UUID, sessionAvgMs int, correctCount int, bestReactionMs int) error {
	start := time.Now()
	query := `UPDATE users SET
	          avg_reaction_ms = CASE
	            WHEN games_played = 0 THEN $2
	            ELSE (avg_reaction_ms * games_played + $2) / (games_played + 1)
	          END,
	          total_correct = total_correct + $3,
	          games_played = games_played + 1,
	          best_reaction_ms = CASE
	            WHEN best_reaction_ms = 0 OR ($4 > 0 AND $4 < best_reaction_ms) THEN $4
	            ELSE best_reaction_ms
	          END
	          WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID, sessionAvgMs, correctCount, bestReactionMs)

	duration := time.Since(start)
	if err != nil {
		r.log.Error("update user stats failed", "op", "UpdateStats", "user_id", userID, "session_avg_ms", sessionAvgMs, "err", err, "duration_ms", duration.Milliseconds(), r.reqIDAttr(ctx))
		return err
	}
	r.warnSlow("UpdateStats", duration, ctx)
	return nil
}

func (r *UserRepository) GetLeaderboard(ctx context.Context, limit int, offset int) ([]model.LeaderboardEntry, int, error) {
	start := time.Now()

	// Single query: window function avoids a separate COUNT(*) round-trip.
	query := `SELECT id, username, avg_reaction_ms, total_correct, games_played, best_reaction_ms,
	                 COUNT(*) OVER() AS total_count
	          FROM users WHERE games_played > 0 AND avg_reaction_ms > 0
	          ORDER BY avg_reaction_ms ASC LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		duration := time.Since(start)
		r.log.Error("get leaderboard failed", "op", "GetLeaderboard", "limit", limit, "offset", offset, "err", err, "duration_ms", duration.Milliseconds(), r.reqIDAttr(ctx))
		return nil, 0, err
	}
	defer rows.Close()

	var totalCount int
	entries := make([]model.LeaderboardEntry, 0)
	rank := offset + 1
	for rows.Next() {
		var e model.LeaderboardEntry
		if err := rows.Scan(&e.UserID, &e.Username, &e.AvgReactionMs, &e.TotalCorrect, &e.GamesPlayed, &e.BestReactionMs, &totalCount); err != nil {
			r.log.Error("scan leaderboard row failed", "op", "GetLeaderboard", "rank", rank, "err", err, r.reqIDAttr(ctx))
			return nil, 0, err
		}
		e.Rank = rank
		rank++
		entries = append(entries, e)
	}

	duration := time.Since(start)
	r.warnSlow("GetLeaderboard", duration, ctx)
	return entries, totalCount, rows.Err()
}

// warnSlow logs a warning if the query took longer than the threshold.
func (r *UserRepository) warnSlow(op string, duration time.Duration, ctx context.Context) {
	if duration > slowQueryThreshold {
		r.log.Warn("slow query", "op", op, "duration_ms", duration.Milliseconds(), r.reqIDAttr(ctx))
	}
}

// reqIDAttr extracts request_id from context for log correlation.
func (r *UserRepository) reqIDAttr(ctx context.Context) slog.Attr {
	if id, ok := ctx.Value(requestIDCtxKey).(string); ok {
		return slog.String("request_id", id)
	}
	return slog.String("request_id", "")
}

// contextKey mirrors the key used in middleware for request_id propagation.
type contextKey string

const requestIDCtxKey contextKey = "request_id"
