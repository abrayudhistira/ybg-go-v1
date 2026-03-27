package http

import (
	"net/http"
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
func (h *PointHandler) GetHistory(c *gin.Context) {
	// 1. Ambil userID dari middleware Auth
	uidObj, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Pastikan casting tipe data sesuai (uuid.UUID)
	uid, ok := uidObj.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	// 2. Panggil Usecase dengan filter UID
	res, err := h.uc.GetMyPointHistory(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal ambil history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": res})
}
func (h *PointHandler) CreatePoint(c *gin.Context) {
	var req struct {
		UserID uuid.UUID `json:"user_id" binding:"required"`
		Point  int       `json:"point" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.uc.AddPointTransaction(req.UserID, req.Point); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update point"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Point updated successfully"})
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
