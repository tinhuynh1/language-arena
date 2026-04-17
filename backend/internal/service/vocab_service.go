package service

import (
	"context"

	"github.com/michael/language-arena/backend/internal/model"
	"github.com/michael/language-arena/backend/internal/repository"
)

type VocabService struct {
	vocabRepo *repository.VocabRepository
}

func NewVocabService(vocabRepo *repository.VocabRepository) *VocabService {
	return &VocabService{vocabRepo: vocabRepo}
}

func (s *VocabService) GetByLanguage(ctx context.Context, q model.VocabQuery) ([]model.Vocabulary, error) {
	if q.Limit <= 0 || q.Limit > 100 {
		q.Limit = 20
	}
	return s.vocabRepo.FindByLanguage(ctx, q)
}

func (s *VocabService) GetRandomSet(ctx context.Context, language, level string, count int) ([]model.Vocabulary, error) {
	if count <= 0 || count > 50 {
		count = 10
	}
	return s.vocabRepo.GetRandomSet(ctx, language, level, count)
}
