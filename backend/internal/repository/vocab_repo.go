package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/michael/language-arena/backend/internal/model"
)

type VocabRepository struct {
	db  *sql.DB
	log *slog.Logger
}

func NewVocabRepository(db *sql.DB) *VocabRepository {
	return &VocabRepository{
		db:  db,
		log: slog.Default().With("component", "REPO.Vocab"),
	}
}

func (r *VocabRepository) FindByLanguage(ctx context.Context, q model.VocabQuery) ([]model.Vocabulary, error) {
	start := time.Now()
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
		duration := time.Since(start)
		r.log.Error("find vocabs by language failed", "op", "FindByLanguage", "language", q.Language, "level", q.Level, "limit", q.Limit, "err", err, "duration_ms", duration.Milliseconds())
		return nil, err
	}
	defer rows.Close()

	vocabs := make([]model.Vocabulary, 0)
	for rows.Next() {
		var v model.Vocabulary
		if err := rows.Scan(&v.ID, &v.Word, &v.Meaning, &v.Language, &v.Level, &v.Difficulty, &v.Category, &v.IPA, &v.Pinyin); err != nil {
			r.log.Error("scan vocab row failed", "op", "FindByLanguage", "err", err)
			return nil, err
		}
		vocabs = append(vocabs, v)
	}

	duration := time.Since(start)
	if duration > slowQueryThreshold {
		r.log.Warn("slow query", "op", "FindByLanguage", "language", q.Language, "level", q.Level, "duration_ms", duration.Milliseconds())
	}
	return vocabs, rows.Err()
}

func (r *VocabRepository) GetRandomSet(ctx context.Context, language, level string, count int) ([]model.Vocabulary, error) {
	start := time.Now()
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
		duration := time.Since(start)
		r.log.Error("get random vocab set failed", "op", "GetRandomSet", "language", language, "level", level, "count", count, "err", err, "duration_ms", duration.Milliseconds())
		return nil, err
	}
	defer rows.Close()

	vocabs := make([]model.Vocabulary, 0)
	for rows.Next() {
		var v model.Vocabulary
		if err := rows.Scan(&v.ID, &v.Word, &v.Meaning, &v.Language, &v.Level, &v.Difficulty, &v.Category, &v.IPA, &v.Pinyin); err != nil {
			r.log.Error("scan vocab row failed", "op", "GetRandomSet", "err", err)
			return nil, err
		}
		vocabs = append(vocabs, v)
	}

	duration := time.Since(start)
	if duration > slowQueryThreshold {
		r.log.Warn("slow query", "op", "GetRandomSet", "language", language, "level", level, "duration_ms", duration.Milliseconds())
	}
	return vocabs, rows.Err()
}
