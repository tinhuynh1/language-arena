package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/michael/language-arena/backend/internal/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (id, username, email, password_hash) 
	          VALUES ($1, $2, $3, $4) RETURNING created_at`
	user.ID = uuid.New()
	return r.db.QueryRowContext(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash,
	).Scan(&user.CreatedAt)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, username, email, password_hash, total_score, games_played, best_reaction_ms, created_at 
	          FROM users WHERE email = $1`
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.TotalScore, &user.GamesPlayed, &user.BestReactionMs, &user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, username, email, password_hash, total_score, games_played, best_reaction_ms, created_at 
	          FROM users WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.TotalScore, &user.GamesPlayed, &user.BestReactionMs, &user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) UpdateStats(ctx context.Context, userID uuid.UUID, scoreAdd int64, reactionMs int) error {
	query := `UPDATE users SET 
	          total_score = total_score + $2, 
	          games_played = games_played + 1,
	          best_reaction_ms = CASE 
	            WHEN best_reaction_ms = 0 OR $3 < best_reaction_ms THEN $3 
	            ELSE best_reaction_ms 
	          END
	          WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID, scoreAdd, reactionMs)
	return err
}

func (r *UserRepository) GetLeaderboard(ctx context.Context, limit int) ([]model.LeaderboardEntry, error) {
	query := `SELECT id, username, total_score, games_played, best_reaction_ms 
	          FROM users ORDER BY total_score DESC LIMIT $1`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries = make([]model.LeaderboardEntry, 0)
	rank := 1
	for rows.Next() {
		var e model.LeaderboardEntry
		if err := rows.Scan(&e.UserID, &e.Username, &e.TotalScore, &e.GamesPlayed, &e.BestReactionMs); err != nil {
			return nil, err
		}
		e.Rank = rank
		rank++
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
