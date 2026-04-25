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
	query := `SELECT id, word, meaning, COALESCE(definition,''), language, level, difficulty, category, COALESCE(ipa,''), COALESCE(pinyin,'') FROM vocabularies
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
		if err := rows.Scan(&v.ID, &v.Word, &v.Meaning, &v.Definition, &v.Language, &v.Level, &v.Difficulty, &v.Category, &v.IPA, &v.Pinyin); err != nil {
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
	query := `SELECT id, word, meaning, COALESCE(definition,''), language, level, difficulty, category, COALESCE(ipa,''), COALESCE(pinyin,'') FROM vocabularies
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
		if err := rows.Scan(&v.ID, &v.Word, &v.Meaning, &v.Definition, &v.Language, &v.Level, &v.Difficulty, &v.Category, &v.IPA, &v.Pinyin); err != nil {
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

func (r *VocabRepository) RecordMistake(ctx context.Context, userID, vocabID, quizType string) error {
	query := `
		INSERT INTO user_mistakes (user_id, vocab_id, quiz_type, incorrect_count, last_mistake_at)
		VALUES ($1, $2, $3, 1, NOW())
		ON CONFLICT (user_id, vocab_id, quiz_type) 
		DO UPDATE SET 
			incorrect_count = user_mistakes.incorrect_count + 1,
			last_mistake_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query, userID, vocabID, quizType)
	if err != nil {
		r.log.Error("failed to record mistake", "op", "RecordMistake", "user_id", userID, "vocab_id", vocabID, "quiz_type", quizType, "err", err)
	}
	return err
}

func (r *VocabRepository) RecordCorrect(ctx context.Context, userID, vocabID, quizType string) error {
	query := `
		INSERT INTO user_mistakes (user_id, vocab_id, quiz_type, correct_count)
		VALUES ($1, $2, $3, 1)
		ON CONFLICT (user_id, vocab_id, quiz_type) 
		DO UPDATE SET correct_count = user_mistakes.correct_count + 1
	`
	// Ignore errors for constraint violations if vocab got deleted etc.
	_, err := r.db.ExecContext(ctx, query, userID, vocabID, quizType)
	if err != nil {
		r.log.Error("failed to record correct answer", "op", "RecordCorrect", "user_id", userID, "vocab_id", vocabID, "err", err)
	}
	return err
}

func (r *VocabRepository) GetTargetedSet(ctx context.Context, userID, language, level, quizType string, count int) ([]model.Vocabulary, error) {
	targetCount := count / 2
	if targetCount == 0 {
		targetCount = 1
	}

	// 1. Fetch mistakes
	query1 := `
		SELECT v.id, v.word, v.meaning, COALESCE(v.definition,''), v.language, v.level, v.difficulty, v.category, COALESCE(v.ipa,''), COALESCE(v.pinyin,'')
		FROM vocabularies v
		INNER JOIN user_mistakes um ON v.id = um.vocab_id
		WHERE um.user_id = $1 AND um.quiz_type = $2 AND v.language = $3 AND um.incorrect_count > um.correct_count
	`
	args1 := []interface{}{userID, quizType, language}
	argIdx := 4

	if level != "" {
		query1 += fmt.Sprintf(` AND v.level = $%d`, argIdx)
		args1 = append(args1, level)
		argIdx++
	}

	query1 += fmt.Sprintf(` ORDER BY RANDOM() LIMIT $%d`, argIdx)
	args1 = append(args1, targetCount)

	rows1, err := r.db.QueryContext(ctx, query1, args1...)
	if err != nil {
		r.log.Error("failed to get mistakes", "err", err)
		return nil, err
	}
	defer func() {
		_ = rows1.Close()
	}()

	mistakeVocabs := make([]model.Vocabulary, 0)
	mistakeIDs := make(map[string]bool)

	for rows1.Next() {
		var v model.Vocabulary
		if err := rows1.Scan(&v.ID, &v.Word, &v.Meaning, &v.Definition, &v.Language, &v.Level, &v.Difficulty, &v.Category, &v.IPA, &v.Pinyin); err == nil {
			mistakeVocabs = append(mistakeVocabs, v)
			mistakeIDs[v.ID.String()] = true
		}
	}

	// 2. Fetch random for the rest
	remaining := count - len(mistakeVocabs)
	if remaining <= 0 {
		return mistakeVocabs, nil
	}

	// Fetch slightly more to account for overlap
	randomVocabs, err := r.GetRandomSet(ctx, language, level, remaining+5)
	if err != nil {
		return mistakeVocabs, nil // return what we have
	}

	// Merge unique
	finalVocabs := mistakeVocabs
	for _, rv := range randomVocabs {
		if len(finalVocabs) >= count {
			break
		}
		if !mistakeIDs[rv.ID.String()] {
			finalVocabs = append(finalVocabs, rv)
			mistakeIDs[rv.ID.String()] = true
		}
	}

	return finalVocabs, nil
}
