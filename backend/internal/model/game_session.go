package model

import (
	"time"

	"github.com/google/uuid"
)

type GameMode string

const (
	ModeSolo   GameMode = "solo"
	ModeDuel   GameMode = "duel"
	ModeBattle GameMode = "battle"
)

type GameSession struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	Mode          GameMode   `json:"mode" db:"mode"`
	Language      string     `json:"language" db:"language"`
	WinnerID      *uuid.UUID `json:"winner_id,omitempty" db:"winner_id"`
	Rounds        int        `json:"rounds" db:"rounds"`
	AvgReactionMs int        `json:"avg_reaction_ms" db:"avg_reaction_ms"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	FinishedAt    *time.Time `json:"finished_at,omitempty" db:"finished_at"`
}

type LeaderboardEntry struct {
	Rank           int       `json:"rank"`
	UserID         uuid.UUID `json:"user_id"`
	Username       string    `json:"username"`
	TotalScore     int64     `json:"total_score"`
	GamesPlayed    int       `json:"games_played"`
	BestReactionMs int       `json:"best_reaction_ms"`
}

type GameSessionPlayer struct {
	ID             uuid.UUID `json:"id" db:"id"`
	SessionID      uuid.UUID `json:"session_id" db:"session_id"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	Score          int       `json:"score" db:"score"`
	AvgReactionMs  int       `json:"avg_reaction_ms" db:"avg_reaction_ms"`
	BestReactionMs int       `json:"best_reaction_ms" db:"best_reaction_ms"`
	Rank           int       `json:"rank" db:"rank"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}
