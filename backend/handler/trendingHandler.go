package handler

import (
	"net/http"
	"second/repository"

	"github.com/gin-gonic/gin"
)

type TrendingHandler struct {
	repo *repository.TrendingRepo
}

func NewTrendingHandler(repo *repository.TrendingRepo) *TrendingHandler {
	return &TrendingHandler{
		repo: repo,
	}
}

func (h *TrendingHandler) GetTrendingDocs(c *gin.Context) {
	res, err := h.repo.GetTrendingDocs()
	if err != nil {
		JSONError(http.StatusInternalServerError, c, err.Error())
	}

	c.JSON(http.StatusOK, res)
}
