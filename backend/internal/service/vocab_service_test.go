package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/michael/language-arena/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Mock ───────────────────────────────────────────────

type mockVocabReader struct {
	findByLanguageFn func(ctx context.Context, q model.VocabQuery) ([]model.Vocabulary, error)
	getRandomSetFn   func(ctx context.Context, language, level string, count int) ([]model.Vocabulary, error)
}

func (m *mockVocabReader) FindByLanguage(ctx context.Context, q model.VocabQuery) ([]model.Vocabulary, error) {
	return m.findByLanguageFn(ctx, q)
}

func (m *mockVocabReader) GetRandomSet(ctx context.Context, language, level string, count int) ([]model.Vocabulary, error) {
	return m.getRandomSetFn(ctx, language, level, count)
}

// ── GetByLanguage Tests ────────────────────────────────

func TestGetByLanguage_Success(t *testing.T) {
	vocabs := []model.Vocabulary{
		{ID: uuid.New(), Word: "hello", Meaning: "xin chào", Language: "en", Level: "A1"},
		{ID: uuid.New(), Word: "world", Meaning: "thế giới", Language: "en", Level: "A1"},
	}
	reader := &mockVocabReader{
		findByLanguageFn: func(_ context.Context, q model.VocabQuery) ([]model.Vocabulary, error) {
			assert.Equal(t, "en", q.Language)
			assert.Equal(t, "A1", q.Level)
			return vocabs, nil
		},
	}

	svc := NewVocabService(reader)
	result, err := svc.GetByLanguage(context.Background(), model.VocabQuery{
		Language: "en",
		Level:    "A1",
		Limit:    20,
	})

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "hello", result[0].Word)
}

func TestGetByLanguage_LimitClamping(t *testing.T) {
	tests := []struct {
		name          string
		inputLimit    int
		expectedLimit int
	}{
		{"zero defaults to 20", 0, 20},
		{"negative defaults to 20", -1, 20},
		{"over 100 defaults to 20", 150, 20},
		{"valid passes through", 50, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &mockVocabReader{
				findByLanguageFn: func(_ context.Context, q model.VocabQuery) ([]model.Vocabulary, error) {
					assert.Equal(t, tt.expectedLimit, q.Limit)
					return nil, nil
				},
			}

			svc := NewVocabService(reader)
			_, err := svc.GetByLanguage(context.Background(), model.VocabQuery{
				Language: "en",
				Limit:    tt.inputLimit,
			})
			require.NoError(t, err)
		})
	}
}

func TestGetByLanguage_RepoError(t *testing.T) {
	dbErr := errors.New("connection refused")
	reader := &mockVocabReader{
		findByLanguageFn: func(_ context.Context, _ model.VocabQuery) ([]model.Vocabulary, error) {
			return nil, dbErr
		},
	}

	svc := NewVocabService(reader)
	result, err := svc.GetByLanguage(context.Background(), model.VocabQuery{Language: "en", Limit: 10})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, dbErr)
}

// ── GetRandomSet Tests ─────────────────────────────────

func TestGetRandomSet_Success(t *testing.T) {
	vocabs := []model.Vocabulary{
		{ID: uuid.New(), Word: "apple", Language: "en", Level: "A1"},
	}
	reader := &mockVocabReader{
		getRandomSetFn: func(_ context.Context, lang, level string, count int) ([]model.Vocabulary, error) {
			assert.Equal(t, "en", lang)
			assert.Equal(t, "A1", level)
			assert.Equal(t, 10, count)
			return vocabs, nil
		},
	}

	svc := NewVocabService(reader)
	result, err := svc.GetRandomSet(context.Background(), "en", "A1", 10)

	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestGetRandomSet_CountClamping(t *testing.T) {
	tests := []struct {
		name          string
		inputCount    int
		expectedCount int
	}{
		{"zero defaults to 10", 0, 10},
		{"negative defaults to 10", -5, 10},
		{"over 50 defaults to 10", 100, 10},
		{"valid passes through", 25, 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &mockVocabReader{
				getRandomSetFn: func(_ context.Context, _, _ string, count int) ([]model.Vocabulary, error) {
					assert.Equal(t, tt.expectedCount, count)
					return nil, nil
				},
			}

			svc := NewVocabService(reader)
			_, err := svc.GetRandomSet(context.Background(), "en", "A1", tt.inputCount)
			require.NoError(t, err)
		})
	}
}

func TestGetRandomSet_RepoError(t *testing.T) {
	reader := &mockVocabReader{
		getRandomSetFn: func(_ context.Context, _, _ string, _ int) ([]model.Vocabulary, error) {
			return nil, errors.New("query failed")
		},
	}

	svc := NewVocabService(reader)
	result, err := svc.GetRandomSet(context.Background(), "zh", "B1", 10)

	assert.Nil(t, result)
	assert.Error(t, err)
}
