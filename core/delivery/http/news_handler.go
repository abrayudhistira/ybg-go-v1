package http

import (
	"net/http"
	"strconv"
	"ybg-backend-go/core/entity"
	"ybg-backend-go/core/usecase"

	"github.com/gin-gonic/gin"
)

type NewsHandler struct {
	uc usecase.NewsUsecase
}

func NewNewsHandler(uc usecase.NewsUsecase) *NewsHandler {
	return &NewsHandler{uc: uc}
}

// Create News (Admin Only via Middleware)
func (h *NewsHandler) Create(c *gin.Context) {
	var n entity.News
	if err := c.ShouldBindJSON(&n); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid", "details": err.Error()})
		return
	}

	if err := h.uc.CreateNews(&n); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat berita"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Berita berhasil dibuat", "data": n})
}

// Get All News (Public/Customer/Admin)
func (h *NewsHandler) GetAll(c *gin.Context) {
	news, err := h.uc.GetAllNews()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": news})
}

// Get News By ID
func (h *NewsHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID berita tidak valid"})
		return
	}

	news, err := h.uc.GetNewsByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Berita tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": news})
}

// Update News (Admin Only via Middleware)
func (h *NewsHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID berita tidak valid"})
		return
	}

	var n entity.News
	if err := c.ShouldBindJSON(&n); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	n.NewsID = uint(id)

	if err := h.uc.UpdateNews(&n); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update berita"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Berita diperbarui", "data": n})
}

// Delete News (Admin Only via Middleware)
func (h *NewsHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID berita tidak valid"})
		return
	}

	if err := h.uc.DeleteNews(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus berita"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Berita berhasil dihapus"})
}
