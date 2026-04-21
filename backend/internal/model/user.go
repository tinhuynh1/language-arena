package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `json:"id" db:"id"`
	Username       string    `json:"username" db:"username"`
	Email          string    `json:"email" db:"email"`
	PasswordHash   string    `json:"-" db:"password_hash"`
	AvgReactionMs  int       `json:"avg_reaction_ms" db:"avg_reaction_ms"`
	TotalCorrect   int       `json:"total_correct" db:"total_correct"`
	GamesPlayed    int       `json:"games_played" db:"games_played"`
	BestReactionMs int       `json:"best_reaction_ms" db:"best_reaction_ms"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
