package http

import (
	"io"
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

func (h *NewsHandler) Create(c *gin.Context) {
	// 1. Ambil data teks dari Form
	n := entity.News{
		Title:       c.PostForm("title"),
		Description: c.PostForm("description"),
		Status:      c.PostForm("status"),
	}

	// Default IsActive ke true jika tidak dikirim
	n.IsActive = true

	// 2. Ambil file gambar
	var imageStream io.Reader
	var fileName, contentType string
	file, err := c.FormFile("image")
	if err == nil {
		openedFile, _ := file.Open()
		defer openedFile.Close()
		imageStream = openedFile
		fileName = file.Filename
		contentType = file.Header.Get("Content-Type")
	}

	if err := h.uc.CreateNews(&n, imageStream, fileName, contentType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat berita", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Berita berhasil dibuat", "data": n})
}

func (h *NewsHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID berita tidak valid"})
		return
	}

	// 1. Ambil data teks
	n := entity.News{
		NewsID:      uint(id),
		Title:       c.PostForm("title"),
		Description: c.PostForm("description"),
		Status:      c.PostForm("status"),
	}

	// 2. Ambil file gambar (hanya jika ada file baru yang diupload)
	var imageStream io.Reader
	var fileName, contentType string
	file, err := c.FormFile("image")
	if err == nil {
		openedFile, _ := file.Open()
		defer openedFile.Close()
		imageStream = openedFile
		fileName = file.Filename
		contentType = file.Header.Get("Content-Type")
	}

	if err := h.uc.UpdateNews(&n, imageStream, fileName, contentType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update berita"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Berita diperbarui", "data": n})
}

// Fungsi GetAll, GetByID, dan Delete tetap menggunakan JSON response standar
func (h *NewsHandler) GetAll(c *gin.Context) {
	news, err := h.uc.GetAllNews()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": news})
}

func (h *NewsHandler) GetByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	news, err := h.uc.GetNewsByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Berita tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": news})
}

func (h *NewsHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.uc.DeleteNews(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus berita"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Berita berhasil dihapus"})
}
