package http

import (
	"io"
	"net/http"
	"strconv"
	"ybg-backend-go/core/entity"
	"ybg-backend-go/core/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RewardHandler struct {
	uc usecase.RewardUsecase
}

func NewRewardHandler(uc usecase.RewardUsecase) *RewardHandler {
	return &RewardHandler{uc: uc}
}

// GET /api/rewards
// Mengambil semua daftar hadiah yang tersedia
func (h *RewardHandler) GetAll(c *gin.Context) {
	rewards, err := h.uc.GetAllRewards()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar reward: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, rewards)
}

// POST /api/rewards/claim
// Melakukan penukaran poin untuk reward tertentu
func (h *RewardHandler) Claim(c *gin.Context) {
	var input struct {
		RewardID string `json:"reward_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reward ID wajib diisi"})
		return
	}

	// Ambil userID dari context (diset oleh AuthMiddleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}

	uID, _ := uuid.Parse(userIDStr.(string))
	rID, err := uuid.Parse(input.RewardID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format Reward ID tidak valid"})
		return
	}

	if err := h.uc.ClaimReward(uID, rID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Klaim berhasil diajukan! Tunggu konfirmasi admin.",
	})
}

// GET /api/rewards/my-history
// Melihat riwayat klaim reward user yang sedang login
func (h *RewardHandler) GetMyHistory(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	uID, _ := uuid.Parse(userIDStr.(string))

	history, err := h.uc.GetMyHistory(uID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil riwayat: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}
func (h *RewardHandler) Approve(c *gin.Context) {
	var input struct {
		HistoryID string `json:"history_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "History ID wajib diisi"})
		return
	}

	hID, err := uuid.Parse(input.HistoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format History ID tidak valid"})
		return
	}

	if err := h.uc.ApproveClaim(hID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Klaim berhasil disetujui, poin user telah dipotong.",
	})
}
func (h *RewardHandler) Reject(c *gin.Context) {
	var input struct {
		HistoryID string `json:"history_id" binding:"required"`
		Reason    string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "History ID wajib diisi"})
		return
	}

	hID, err := uuid.Parse(input.HistoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format History ID tidak valid"})
		return
	}

	if err := h.uc.RejectClaim(hID, input.Reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Klaim telah ditolak.",
	})
}
func (h *RewardHandler) Create(c *gin.Context) {
	// 1. Parsing data teks dari Form
	pointCost, _ := strconv.Atoi(c.PostForm("point_cost"))
	quantity, _ := strconv.Atoi(c.PostForm("quantity"))

	reward := &entity.Reward{
		Name:        c.PostForm("name"),
		Description: c.PostForm("description"),
		PointCost:   pointCost,
		Quantity:    quantity,
		Category:    c.PostForm("category"),
	}

	// 2. Ambil file gambar (Opsional)
	file, err := c.FormFile("image")
	var img io.Reader
	var fileName, contentType string

	if err == nil {
		// Validasi ukuran (misal 2MB)
		if file.Size > 2*1024*1024 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File terlalu besar (max 2MB)"})
			return
		}

		openedFile, _ := file.Open()
		defer openedFile.Close()
		img = openedFile
		fileName = file.Filename
		contentType = file.Header.Get("Content-Type")
	}

	// 3. Panggil Usecase
	if err := h.uc.CreateReward(reward, img, fileName, contentType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Reward berhasil dibuat",
		"data":    reward,
	})
}

// PUT /api/rewards/admin/:id
func (h *RewardHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	rID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format ID tidak valid"})
		return
	}

	pointCost, _ := strconv.Atoi(c.PostForm("point_cost"))
	quantity, _ := strconv.Atoi(c.PostForm("quantity"))

	reward := &entity.Reward{
		RewardID:    rID,
		Name:        c.PostForm("name"),
		Description: c.PostForm("description"),
		PointCost:   pointCost,
		Quantity:    quantity,
		Category:    c.PostForm("category"),
	}

	// Handle Image Opsional
	file, err := c.FormFile("image")
	var img io.Reader
	var fileName, contentType string
	if err == nil {
		openedFile, _ := file.Open()
		defer openedFile.Close()
		img = openedFile
		fileName = file.Filename
		contentType = file.Header.Get("Content-Type")
	}

	if err := h.uc.UpdateReward(reward, img, fileName, contentType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reward berhasil diperbarui", "data": reward})
}

// DELETE /api/rewards/admin/:id
func (h *RewardHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	rID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format ID tidak valid"})
		return
	}

	if err := h.uc.DeleteReward(rID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reward berhasil dihapus"})
}
