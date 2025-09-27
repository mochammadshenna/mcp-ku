package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"mcp-octo-enigma/internal/core/services"
)

type SearchHandler struct { rag *services.RAGService }

func NewSearchHandler(rag *services.RAGService) *SearchHandler { return &SearchHandler{rag: rag} }

func (h *SearchHandler) Register(r *gin.Engine) {
	r.GET("/api/v1/search", h.search)
}

func (h *SearchHandler) search(c *gin.Context) {
	q := c.Query("q")
	if q == "" { c.JSON(http.StatusBadRequest, gin.H{"error": "q is required"}); return }
	topK, _ := strconv.Atoi(c.Query("k"))
	if topK <= 0 { topK = 5 }
	res, err := h.rag.Query(c, q, topK)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
	c.JSON(http.StatusOK, res)
}
