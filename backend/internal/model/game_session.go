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
	Player1ID     uuid.UUID  `json:"player1_id" db:"player1_id"`
	Player2ID     *uuid.UUID `json:"player2_id,omitempty" db:"player2_id"`
	Player1Score  int        `json:"player1_score" db:"player1_score"`
	Player2Score  int        `json:"player2_score" db:"player2_score"`
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
