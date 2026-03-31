package http

import (
	"net/http"
	"strconv"
	"ybg-backend-go/core/usecase"

	"github.com/gin-gonic/gin"
)

type CartHandler struct {
	uc usecase.CartUsecase
}

func NewCartHandler(uc usecase.CartUsecase) *CartHandler {
	return &CartHandler{uc: uc}
}

func (h *CartHandler) GetMyCart(c *gin.Context) {
	userID := c.GetString("user_id") // Pastikan middleware Auth nyimpen ini
	cart, err := h.uc.GetCart(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Keranjang kamu masih kosong"})
		return
	}
	c.JSON(http.StatusOK, cart)
}

func (h *CartHandler) AddToCart(c *gin.Context) {
	userID := c.GetString("user_id")
	var input struct {
		ProductID uint `json:"product_id" binding:"required"`
		Quantity  int  `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid"})
		return
	}

	if err := h.uc.AddItemToCart(userID, input.ProductID, input.Quantity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Produk berhasil ditambahkan ke keranjang"})
}

func (h *CartHandler) DeleteItem(c *gin.Context) {
	userID := c.GetString("user_id")

	// SESUAIKAN: Jika di index.go kamu tulis "/:id", maka pakai c.Param("id")
	// Jika di index.go kamu tulis "/:product_id", maka pakai c.Param("product_id")
	paramID := c.Param("id")
	if paramID == "" {
		paramID = c.Param("product_id") // Fallback jika namanya berbeda
	}

	// Konversi string ke uint
	productID64, err := strconv.ParseUint(paramID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID produk tidak valid"})
		return
	}
	productID := uint(productID64)

	// Panggil Usecase
	if err := h.uc.RemoveItemFromCart(userID, productID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Produk berhasil dihapus dari keranjang"})
}
func (h *CartHandler) ClearMyCart(c *gin.Context) {
	userID := c.GetString("user_id")

	if err := h.uc.ClearCart(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Keranjang berhasil dikosongkan"})
}
