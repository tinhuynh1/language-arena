package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/michael/language-arena/backend/internal/model"
)

type VocabRepository struct {
	db *sql.DB
}

func NewVocabRepository(db *sql.DB) *VocabRepository {
	return &VocabRepository{db: db}
}

func (r *VocabRepository) FindByLanguage(ctx context.Context, q model.VocabQuery) ([]model.Vocabulary, error) {
	query := `SELECT id, word, meaning, language, level, difficulty, category, COALESCE(ipa,''), COALESCE(pinyin,'') FROM vocabularies
	          WHERE language = $1`
	args := []interface{}{q.Language}
	argIdx := 2

	if q.Level != "" {
		query += fmt.Sprintf(` AND level = $%d`, argIdx)
		args = append(args, q.Level)
		argIdx++
	}

	query += fmt.Sprintf(` ORDER BY RANDOM() LIMIT $%d`, argIdx)
	args = append(args, q.Limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vocabs := make([]model.Vocabulary, 0)
	for rows.Next() {
		var v model.Vocabulary
		if err := rows.Scan(&v.ID, &v.Word, &v.Meaning, &v.Language, &v.Level, &v.Difficulty, &v.Category, &v.IPA, &v.Pinyin); err != nil {
			return nil, err
		}
		vocabs = append(vocabs, v)
	}
	return vocabs, rows.Err()
}

func (r *VocabRepository) GetRandomSet(ctx context.Context, language, level string, count int) ([]model.Vocabulary, error) {
	query := `SELECT id, word, meaning, language, level, difficulty, category, COALESCE(ipa,''), COALESCE(pinyin,'') FROM vocabularies
	          WHERE language = $1`
	args := []interface{}{language}
	argIdx := 2

	if level != "" {
		query += fmt.Sprintf(` AND level = $%d`, argIdx)
		args = append(args, level)
		argIdx++
	}

	query += fmt.Sprintf(` ORDER BY RANDOM() LIMIT $%d`, argIdx)
	args = append(args, count)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vocabs := make([]model.Vocabulary, 0)
	for rows.Next() {
		var v model.Vocabulary
		if err := rows.Scan(&v.ID, &v.Word, &v.Meaning, &v.Language, &v.Level, &v.Difficulty, &v.Category, &v.IPA, &v.Pinyin); err != nil {
			return nil, err
		}
		vocabs = append(vocabs, v)
	}
	return vocabs, rows.Err()
}
