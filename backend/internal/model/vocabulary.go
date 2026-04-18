package model

import "github.com/google/uuid"

type Vocabulary struct {
	ID         uuid.UUID `json:"id" db:"id"`
	Word       string    `json:"word" db:"word"`
	Meaning    string    `json:"meaning" db:"meaning"`
	Language   string    `json:"language" db:"language"`
	Level      string    `json:"level" db:"level"`
	Difficulty int       `json:"difficulty" db:"difficulty"`
	Category   string    `json:"category" db:"category"`
	IPA        string    `json:"ipa" db:"ipa"`
	Pinyin     string    `json:"pinyin" db:"pinyin"`
}

type VocabQuery struct {
	Language string `form:"lang" binding:"required,oneof=en zh"`
	Level    string `form:"level"`
	Limit    int    `form:"limit,default=20"`
}
