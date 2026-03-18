package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"swd-new/internal/service"
	"swd-new/pkg/helper/resp"
)

type SensitiveCheckHandler struct {
	*Handler
	sensitiveWordService service.SensitiveWordService
}

type checkRequest struct {
	Text string `json:"text" binding:"required"`
}

type createSensitiveWordRequest struct {
	Word string `json:"word" binding:"required"`
	Type string `json:"type"`
}

type updateSensitiveWordRequest struct {
	Word string `json:"word" binding:"required"`
	Type string `json:"type"`
}

func NewSensitiveCheckHandler(handler *Handler, sensitiveWordService service.SensitiveWordService) *SensitiveCheckHandler {
	return &SensitiveCheckHandler{
		Handler:              handler,
		sensitiveWordService: sensitiveWordService,
	}
}

func (h *SensitiveCheckHandler) Check(ctx *gin.Context) {
	var req checkRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.HandleError(ctx, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	result, err := h.sensitiveWordService.Check(req.Text)
	if err != nil {
		resp.HandleError(ctx, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	resp.HandleSuccess(ctx, result)
}

func (h *SensitiveCheckHandler) ListWords(ctx *gin.Context) {
	pageNum, _ := strconv.Atoi(ctx.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	result, err := h.sensitiveWordService.ListWords(ctx.Request.Context(), pageNum, pageSize)
	if err != nil {
		resp.HandleError(ctx, http.StatusInternalServerError, 500, err.Error(), nil)
		return
	}

	resp.HandleSuccess(ctx, result)
}

func (h *SensitiveCheckHandler) CreateWord(ctx *gin.Context) {
	var req createSensitiveWordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.HandleError(ctx, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	word, err := h.sensitiveWordService.CreateWord(ctx.Request.Context(), service.CreateSensitiveWordInput{
		Word: strings.TrimSpace(req.Word),
		Type: strings.TrimSpace(req.Type),
	})
	if err != nil {
		resp.HandleError(ctx, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	resp.HandleSuccess(ctx, word)
}

func (h *SensitiveCheckHandler) UpdateWord(ctx *gin.Context) {
	id, err := parseSensitiveWordID(ctx.Param("id"))
	if err != nil {
		resp.HandleError(ctx, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	var req updateSensitiveWordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.HandleError(ctx, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	word, err := h.sensitiveWordService.UpdateWord(ctx.Request.Context(), id, service.UpdateSensitiveWordInput{
		Word: strings.TrimSpace(req.Word),
		Type: strings.TrimSpace(req.Type),
	})
	if err != nil {
		resp.HandleError(ctx, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	resp.HandleSuccess(ctx, word)
}

func (h *SensitiveCheckHandler) DeleteWord(ctx *gin.Context) {
	id, err := parseSensitiveWordID(ctx.Param("id"))
	if err != nil {
		resp.HandleError(ctx, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	if err := h.sensitiveWordService.DeleteWord(ctx.Request.Context(), id); err != nil {
		resp.HandleError(ctx, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	resp.HandleSuccess(ctx, gin.H{"id": id})
}

func parseSensitiveWordID(value string) (uint, error) {
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, service.ErrInvalidSensitiveWordID
	}
	return uint(id), nil
}
