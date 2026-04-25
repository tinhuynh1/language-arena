package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/michael/language-arena/backend/internal/model"
	"github.com/michael/language-arena/backend/internal/service"
	"github.com/stretchr/testify/assert"
)

// ── Mock ───────────────────────────────────────────────

type mockVocabReaderHandler struct {
	findByLanguageFn func(ctx context.Context, q model.VocabQuery) ([]model.Vocabulary, error)
	getRandomSetFn   func(ctx context.Context, language, level string, count int) ([]model.Vocabulary, error)
}

func (m *mockVocabReaderHandler) FindByLanguage(ctx context.Context, q model.VocabQuery) ([]model.Vocabulary, error) {
	return m.findByLanguageFn(ctx, q)
}

func (m *mockVocabReaderHandler) GetRandomSet(ctx context.Context, language, level string, count int) ([]model.Vocabulary, error) {
	if m.getRandomSetFn != nil {
		return m.getRandomSetFn(ctx, language, level, count)
	}
	return nil, nil
}

func (m *mockVocabReaderHandler) GetTargetedSet(ctx context.Context, userID, language, level, quizType string, count int) ([]model.Vocabulary, error) {
	return nil, nil
}

func (m *mockVocabReaderHandler) RecordMistake(ctx context.Context, userID, vocabID, quizType string) error {
	return nil
}

func (m *mockVocabReaderHandler) RecordCorrect(ctx context.Context, userID, vocabID, quizType string) error {
	return nil
}

// ── GetVocabularies Tests ──────────────────────────────

func TestGetVocabularies_Success(t *testing.T) {
	reader := &mockVocabReaderHandler{
		findByLanguageFn: func(_ context.Context, _ model.VocabQuery) ([]model.Vocabulary, error) {
			return []model.Vocabulary{{Word: "hello"}}, nil
		},
	}

	svc := service.NewVocabService(reader)
	h := NewVocabHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/vocab?lang=en", nil)

	h.GetVocabularies(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetVocabularies_InvalidQuery(t *testing.T) {
	reader := &mockVocabReaderHandler{
		findByLanguageFn: func(_ context.Context, _ model.VocabQuery) ([]model.Vocabulary, error) {
			return nil, nil
		},
	}

	svc := service.NewVocabService(reader)
	h := NewVocabHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// missing required 'lang' parameter
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/vocab", nil)

	h.GetVocabularies(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetVocabularies_InternalError(t *testing.T) {
	reader := &mockVocabReaderHandler{
		findByLanguageFn: func(_ context.Context, _ model.VocabQuery) ([]model.Vocabulary, error) {
			return nil, errors.New("db error")
		},
	}

	svc := service.NewVocabService(reader)
	h := NewVocabHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/vocab?lang=en", nil)

	h.GetVocabularies(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
