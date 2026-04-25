package service

import (
	"context"
	"log/slog"

	"github.com/michael/language-arena/backend/internal/model"
)

// VocabReader abstracts vocabulary reads.
type VocabReader interface {
	FindByLanguage(ctx context.Context, q model.VocabQuery) ([]model.Vocabulary, error)
	GetRandomSet(ctx context.Context, language, level string, count int) ([]model.Vocabulary, error)
}

type VocabService struct {
	vocabReader VocabReader
	log         *slog.Logger
}

func NewVocabService(vocabReader VocabReader) *VocabService {
	return &VocabService{
		vocabReader: vocabReader,
		log:         slog.Default().With("component", "SVC.Vocab"),
	}
}

func (s *VocabService) GetByLanguage(ctx context.Context, q model.VocabQuery) ([]model.Vocabulary, error) {
	if q.Limit <= 0 || q.Limit > 100 {
		q.Limit = 20
	}

	vocabs, err := s.vocabReader.FindByLanguage(ctx, q)
	if err != nil {
		s.log.Error("get vocabs by language failed", "op", "GetByLanguage", "language", q.Language, "level", q.Level, "limit", q.Limit, "err", err)
		return nil, err
	}

	s.log.Debug("vocabs fetched", "op", "GetByLanguage", "language", q.Language, "level", q.Level, "count", len(vocabs))
	return vocabs, nil
}

func (s *VocabService) GetRandomSet(ctx context.Context, language, level string, count int) ([]model.Vocabulary, error) {
	if count <= 0 || count > 50 {
		count = 10
	}

	vocabs, err := s.vocabReader.GetRandomSet(ctx, language, level, count)
	if err != nil {
		s.log.Error("get random vocab set failed", "op", "GetRandomSet", "language", language, "level", level, "count", count, "err", err)
		return nil, err
	}

	s.log.Debug("random vocabs fetched", "op", "GetRandomSet", "language", language, "level", level, "count", len(vocabs))
	return vocabs, nil
}
