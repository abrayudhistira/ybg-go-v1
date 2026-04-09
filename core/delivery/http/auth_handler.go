package http

import (
	"net/http"
	"ybg-backend-go/core/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	uc usecase.AuthUsecase
}

func NewAuthHandler(uc usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email tidak valid"})
		return
	}

	if err := h.uc.RequestOTP(input.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP telah dikirim ke Gmail kamu"})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var input struct {
		Email       string `json:"email" binding:"required"`
		OTP         string `json:"otp" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak lengkap atau password kurang dari 6 karakter"})
		return
	}

	if err := h.uc.VerifyOTPAndResetPassword(input.Email, input.OTP, input.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password berhasil diperbarui, silakan login kembali"})
}

func (h *AuthHandler) RequestChangeEmail(c *gin.Context) {
	var input struct {
		NewEmail string `json:"new_email" binding:"required,email"`
	}

	// 1. Validasi format input
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format email tidak valid"})
		return
	}

	// 2. Panggil Usecase untuk cek email & kirim OTP
	if err := h.uc.RequestEmailUpdateOTP(input.NewEmail); err != nil {
		// Error ini bisa muncul kalau email sudah terdaftar atau gagal kirim Gmail
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3. Respon sukses
	c.JSON(http.StatusOK, gin.H{
		"message": "Kode verifikasi telah dikirim ke email baru anda: " + input.NewEmail,
	})
}
func (h *AuthHandler) VerifyChangeEmail(c *gin.Context) {
	var input struct {
		NewEmail string `json:"new_email" binding:"required,email"`
		OTP      string `json:"otp" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid"})
		return
	}

	// Ambil UserID dari middleware Auth (JWT)
	// Karena hanya user yang login yang boleh ganti emailnya sendiri
	userIDStr := c.MustGet("user_id").(string)
	userID, _ := uuid.Parse(userIDStr)

	if err := h.uc.VerifyAndChangeEmail(userID, input.NewEmail, input.OTP); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email berhasil diperbarui!"})
}

// api/http/auth_handler.go

func (h *AuthHandler) VerifyRegistration(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
		OTP   string `json:"otp" binding:"required"`
	}

	// Validasi JSON body
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email dan OTP wajib diisi dengan format yang benar",
		})
		return
	}

	// Panggil Usecase untuk verifikasi
	if err := h.uc.VerifyAccount(input.Email, input.OTP); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Akun berhasil diverifikasi! Sekarang kamu bisa login.",
	})
}
