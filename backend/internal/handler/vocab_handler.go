package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/michael/language-arena/backend/internal/middleware"
	"github.com/michael/language-arena/backend/internal/model"
	"github.com/michael/language-arena/backend/internal/service"
	"github.com/michael/language-arena/backend/pkg/response"
)

type VocabHandler struct {
	vocabService *service.VocabService
	log          *slog.Logger
}

func NewVocabHandler(vocabService *service.VocabService) *VocabHandler {
	return &VocabHandler{
		vocabService: vocabService,
		log:          slog.Default().With("component", "HANDLER.Vocab"),
	}
}

func (h *VocabHandler) GetVocabularies(c *gin.Context) {
	var q model.VocabQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		h.log.Warn("get vocabularies: invalid query params", "err", err, "request_id", middleware.RequestIDFromContext(c.Request.Context()))
		response.BadRequest(c, "invalid query: lang must be 'en' or 'zh'")
		return
	}

	vocabs, err := h.vocabService.GetByLanguage(c.Request.Context(), q)
	if err != nil {
		h.log.Error("get vocabularies: fetch failed", "language", q.Language, "level", q.Level, "limit", q.Limit, "err", err, "request_id", middleware.RequestIDFromContext(c.Request.Context()))
		response.InternalError(c, "failed to fetch vocabularies")
		return
	}

	response.OK(c, vocabs)
}
