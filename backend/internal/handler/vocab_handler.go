package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/michael/language-arena/backend/internal/model"
	"github.com/michael/language-arena/backend/internal/service"
	"github.com/michael/language-arena/backend/pkg/response"
)

type VocabHandler struct {
	vocabService *service.VocabService
}

func NewVocabHandler(vocabService *service.VocabService) *VocabHandler {
	return &VocabHandler{vocabService: vocabService}
}

func (h *VocabHandler) GetVocabularies(c *gin.Context) {
	var q model.VocabQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.BadRequest(c, "invalid query: lang must be 'en' or 'zh'")
		return
	}

	vocabs, err := h.vocabService.GetByLanguage(c.Request.Context(), q)
	if err != nil {
		response.InternalError(c, "failed to fetch vocabularies")
		return
	}

	response.OK(c, vocabs)
}
