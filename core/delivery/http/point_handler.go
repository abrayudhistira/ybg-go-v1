package http

import (
	"net/http"
	"time"
	"ybg-backend-go/core/entity"
	"ybg-backend-go/core/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PointHandler struct {
	uc usecase.PointUsecase
}

func NewPointHandler(uc usecase.PointUsecase) *PointHandler { return &PointHandler{uc: uc} }

// func (h *PointHandler) GetHistory(c *gin.Context) {
// 	// Ambil userID dari middleware Auth (pastikan di middleware kamu Set("user_id", ...))
// 	uidObj, _ := c.Get("user_id")
// 	uid := uidObj.(uuid.UUID)

//		res, err := h.uc.GetMyPointHistory(uid)
//		if err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal ambil history"})
//			return
//		}
//		c.JSON(http.StatusOK, gin.H{"data": res})
//	}
// func (h *PointHandler) GetHistory(c *gin.Context) {
// 	// 1. Ambil userID dari middleware Auth
// 	uidObj, exists := c.Get("user_id")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
// 		return
// 	}

// 	// Pastikan casting tipe data sesuai (uuid.UUID)
// 	uid, ok := uidObj.(uuid.UUID)
// 	if !ok {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
// 		return
// 	}

// 	// 2. Panggil Usecase dengan filter UID
// 	res, err := h.uc.GetMyPointHistory(uid)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal ambil history"})
// 		return
// 	}

//		c.JSON(http.StatusOK, gin.H{"data": res})
//	}
func (h *PointHandler) GetHistory(c *gin.Context) {
	var targetUID uuid.UUID
	var err error

	// 1. Prioritas: Cek apakah ada user_id di Query Parameter (?user_id=...)
	// Ini digunakan saat Admin sedang mengecek history user tertentu
	queryUID := c.Query("user_id")

	if queryUID != "" {
		targetUID, err = uuid.Parse(queryUID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format User ID tidak valid (harus UUID)"})
			return
		}
	} else {
		// 2. Fallback: Jika tidak ada query param, ambil dari JWT (untuk user melihat dirinya sendiri)
		uidObj, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Sesi tidak ditemukan, silakan login kembali"})
			return
		}

		var ok bool
		targetUID, ok = uidObj.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses identitas user dari token"})
			return
		}
	}

	// 3. Panggil Usecase dengan targetUID yang sudah didapat
	res, err := h.uc.GetMyPointHistory(targetUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil riwayat poin dari database"})
		return
	}

	// 4. Return data (pastikan return array kosong [] jika data tidak ada, jangan nil)
	if res == nil {
		res = []entity.PointHistory{}
	}

	c.JSON(http.StatusOK, gin.H{"data": res})
}
func (h *PointHandler) CreatePoint(c *gin.Context) {
	var req struct {
		UserID    uuid.UUID `json:"user_id" binding:"required"`
		Point     int       `json:"point" binding:"required"`
		ExpiredAt string    `json:"expired_at" binding:"required"` // Format: "2026-12-31T23:59:59Z"
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID, Point, dan Expired At wajib diisi"})
		return
	}

	// 1. Validasi Poin Positif
	if req.Point <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Poin harus lebih besar dari 0"})
		return
	}

	// 2. Parsing String ke Time
	expiryTime, err := time.Parse(time.RFC3339, req.ExpiredAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format tanggal salah. Gunakan format RFC3339 (contoh: 2026-12-31T23:59:59Z)"})
		return
	}

	// 3. Validasi: Tanggal expired tidak boleh di masa lalu
	if expiryTime.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tanggal kadaluwarsa tidak boleh di masa lalu"})
		return
	}

	if err := h.uc.AddPointTransaction(req.UserID, req.Point, expiryTime); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menambah poin"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Poin berhasil ditambahkan dengan masa aktif hingga " + req.ExpiredAt})
}
func (h *PointHandler) GetAllSummaries(c *gin.Context) {
	totals, err := h.uc.FetchAllUsersPoints()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data poin"})
		return
	}

	// Kita buat slice baru untuk format response sesuai mau kamu
	type pointResponse struct {
		UserID uuid.UUID `json:"user_id"`
		Name   string    `json:"name"`
		Point  int       `json:"point"`
	}

	var response []pointResponse
	for _, t := range totals {
		userName := "Unknown"
		if t.User != nil {
			userName = t.User.Name
		}

		response = append(response, pointResponse{
			UserID: t.UserID,
			Name:   userName,
			Point:  t.Total,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}
