package service

import (
	"context"

	"github.com/michael/language-arena/backend/internal/model"
)

// VocabReader abstracts vocabulary reads.
type VocabReader interface {
	FindByLanguage(ctx context.Context, q model.VocabQuery) ([]model.Vocabulary, error)
	GetRandomSet(ctx context.Context, language, level string, count int) ([]model.Vocabulary, error)
}

type VocabService struct {
	vocabReader VocabReader
}

func NewVocabService(vocabReader VocabReader) *VocabService {
	return &VocabService{vocabReader: vocabReader}
}

func (s *VocabService) GetByLanguage(ctx context.Context, q model.VocabQuery) ([]model.Vocabulary, error) {
	if q.Limit <= 0 || q.Limit > 100 {
		q.Limit = 20
	}
	return s.vocabReader.FindByLanguage(ctx, q)
}

func (s *VocabService) GetRandomSet(ctx context.Context, language, level string, count int) ([]model.Vocabulary, error) {
	if count <= 0 || count > 50 {
		count = 10
	}
	return s.vocabReader.GetRandomSet(ctx, language, level, count)
}
