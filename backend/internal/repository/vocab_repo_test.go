package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/michael/language-arena/backend/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestVocabRepository_FindByLanguage(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewVocabRepository(db)
	ctx := context.Background()

	t.Run("success without level", func(t *testing.T) {
		id := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "word", "meaning", "language", "level", "difficulty", "category", "ipa", "pinyin"}).
			AddRow(id, "apple", "quả táo", "en", "A1", 1, "fruit", "æpl", "")

		mock.ExpectQuery("SELECT id, word, meaning, language").WithArgs("en", 10).WillReturnRows(rows)

		res, err := repo.FindByLanguage(ctx, model.VocabQuery{Language: "en", Limit: 10})
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, "apple", res[0].Word)
	})

	t.Run("success with level", func(t *testing.T) {
		id := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "word", "meaning", "language", "level", "difficulty", "category", "ipa", "pinyin"}).
			AddRow(id, "apple", "quả táo", "en", "A1", 1, "fruit", "", "")

		mock.ExpectQuery("SELECT id, word, meaning, language").WithArgs("en", "A1", 10).WillReturnRows(rows)

		res, err := repo.FindByLanguage(ctx, model.VocabQuery{Language: "en", Level: "A1", Limit: 10})
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, word, meaning, language").WithArgs("en", 10).WillReturnError(sql.ErrConnDone)

		res, err := repo.FindByLanguage(ctx, model.VocabQuery{Language: "en", Limit: 10})
		assert.ErrorIs(t, err, sql.ErrConnDone)
		assert.Nil(t, res)
	})
}

func TestVocabRepository_GetRandomSet(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewVocabRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		id := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "word", "meaning", "language", "level", "difficulty", "category", "ipa", "pinyin"}).
			AddRow(id, "dog", "con chó", "en", "A1", 1, "animal", "", "")

		mock.ExpectQuery("SELECT id, word, meaning, language").WithArgs("en", "A1", 5).WillReturnRows(rows)

		res, err := repo.GetRandomSet(ctx, "en", "A1", 5)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, "dog", res[0].Word)
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, word, meaning, language").WithArgs("en", 5).WillReturnError(sql.ErrNoRows)

		res, err := repo.GetRandomSet(ctx, "en", "", 5)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, res)
	})
}
