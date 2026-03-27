package http

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"ybg-backend-go/core/entity"
	"ybg-backend-go/core/usecase"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductHandler struct {
	uc usecase.ProductUsecase
}

func NewProductHandler(uc usecase.ProductUsecase) *ProductHandler {
	return &ProductHandler{uc: uc}
}

func (h *ProductHandler) parseProductForm(c *gin.Context) (*entity.Product, io.Reader, string, string, error) {
	price, _ := strconv.ParseFloat(c.PostForm("price"), 64)
	brandID, _ := strconv.Atoi(c.PostForm("brand_id"))
	categoryID, _ := strconv.Atoi(c.PostForm("category_id"))
	stock, _ := strconv.Atoi(c.PostForm("stock"))

	p := &entity.Product{
		Name:        c.PostForm("name"),
		Price:       price,
		BrandID:     uint(brandID),
		CategoryID:  uint(categoryID),
		Description: c.PostForm("description"),
		Stock:       stock,
		Condition:   c.PostForm("condition"),
		Status:      c.PostForm("status"),
	}

	file, err := c.FormFile("image")
	if err != nil {
		// Gambar bersifat opsional
		return p, nil, "", "", nil
	}

	// Validasi Ukuran: Gunakan pesan string manual agar kompatibel dengan semua versi Go
	if file.Size > 5*1024*1024 {
		return nil, nil, "", "", errors.New("file size exceeds 5MB limit")
	}

	openedFile, err := file.Open()
	if err != nil {
		return nil, nil, "", "", err
	}

	return p, openedFile, file.Filename, file.Header.Get("Content-Type"), nil
}

func (h *ProductHandler) Create(c *gin.Context) {
	p, img, name, cType, err := h.parseProductForm(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Jika ada gambar, pastikan ditutup setelah usecase selesai
	if img != nil {
		if closer, ok := img.(io.Closer); ok {
			defer closer.Close()
		}
	}

	if err := h.uc.CreateProduct(p, img, name, cType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product", "details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Product created", "data": p})
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	p, img, name, cType, err := h.parseProductForm(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p.ProductID = uint(id)

	if img != nil {
		if closer, ok := img.(io.Closer); ok {
			defer closer.Close()
		}
	}

	if err := h.uc.UpdateProduct(p, img, name, cType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product updated", "data": p})
}

// func (h *ProductHandler) GetAll(c *gin.Context) {
// 	products, err := h.uc.FetchProducts()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
// 		return
// 	}

// 	if len(products) == 0 {
// 		c.JSON(http.StatusOK, gin.H{"message": "No products found", "data": []entity.Product{}})
// 		return
// 	}

//		c.JSON(http.StatusOK, gin.H{"data": products})
//	}
func (h *ProductHandler) GetAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	products, total, err := h.uc.FetchProducts("", limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": products,
		"meta": gin.H{"total": total, "page": page, "has_more": int64(offset+limit) < total},
	})
}

// SEARCH ENDPOINT
func (h *ProductHandler) Search(c *gin.Context) {
	query := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	products, total, err := h.uc.FetchProducts(query, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": products,
		"meta": gin.H{"total": total, "page": page, "has_more": int64(offset+limit) < total},
	})
}
func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	product, err := h.uc.GetProductDetail(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": product})
}

func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.uc.DeleteProduct(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
