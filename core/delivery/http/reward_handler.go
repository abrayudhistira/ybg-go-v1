package http

import (
	"net/http"
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
	var input entity.Reward
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid: " + err.Error()})
		return
	}

	if err := h.uc.CreateReward(&input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Reward berhasil ditambahkan ke katalog",
		"data":    input,
	})
}
