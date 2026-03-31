package http

import (
	"io"
	"net/http"
	"strconv"
	"ybg-backend-go/core/entity"
	"ybg-backend-go/core/usecase"

	"github.com/gin-gonic/gin"
)

type BrandHandler struct {
	uc usecase.BrandUsecase
}

func NewBrandHandler(uc usecase.BrandUsecase) *BrandHandler { return &BrandHandler{uc: uc} }

//	func (h *BrandHandler) Create(c *gin.Context) {
//		var b entity.Brand
//		if err := c.ShouldBind(&b); err != nil {
//			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//			return
//		}
//		if err := h.uc.CreateBrand(&b); err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//			return
//		}
//		c.JSON(http.StatusCreated, gin.H{"data": b})
//	}
func (h *BrandHandler) Create(c *gin.Context) {
	// Ambil Nama Brand
	name := c.PostForm("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	// Ambil File Gambar
	fileHeader, err := c.FormFile("image") // Key: "image"
	var file io.Reader
	var fileName, contentType string

	if err == nil {
		openedFile, _ := fileHeader.Open()
		defer openedFile.Close()
		file = openedFile
		fileName = fileHeader.Filename
		contentType = fileHeader.Header.Get("Content-Type")
	}

	brand := entity.Brand{
		Name: name,
	}

	// Kirim ke Usecase
	if err := h.uc.CreateBrand(&brand, file, fileName, contentType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": brand})
}
func (h *BrandHandler) GetAll(c *gin.Context) {
	brands, _ := h.uc.GetAllBrands()
	c.JSON(http.StatusOK, gin.H{"data": brands})
}

func (h *BrandHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	h.uc.DeleteBrand(uint(id))
	c.JSON(http.StatusOK, gin.H{"message": "Brand deleted"})
}
