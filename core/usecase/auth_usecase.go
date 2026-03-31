package usecase

import (
	"crypto/rand"
	"errors"
	"io"
	"time"
	"ybg-backend-go/core/entity"
	"ybg-backend-go/core/repository"
	"ybg-backend-go/pkg/utils"
)

type AuthUsecase interface {
	RequestOTP(email string) error
	VerifyOTPAndResetPassword(email, otp, newPassword string) error
}

type authUC struct {
	userRepo repository.UserRepository
	authRepo repository.AuthRepository
}

func NewAuthUsecase(u repository.UserRepository, a repository.AuthRepository) AuthUsecase {
	return &authUC{userRepo: u, authRepo: a}
}

// 1. Generate OTP & Kirim Email
func (u *authUC) RequestOTP(email string) error {
	// Cek apakah user ada di database
	_, err := u.userRepo.GetByEmail(email)
	if err != nil {
		return errors.New("user dengan email tersebut tidak ditemukan")
	}

	// Generate 6 digit angka random
	otp := u.generateRandomOTP(6)

	// Simpan ke DB dengan expiry 15 menit
	resetData := &entity.PasswordReset{
		Email:     email,
		OTP:       otp,
		ExpiredAt: time.Now().Add(15 * time.Minute),
	}

	if err := u.authRepo.SaveOTP(resetData); err != nil {
		return err
	}

	// Kirim via Gmail
	return utils.SendOTPEmail(email, otp)
}

// 2. Verifikasi OTP & Update Password
func (u *authUC) VerifyOTPAndResetPassword(email, otp, newPassword string) error {
	// Ambil data OTP dari DB
	reset, err := u.authRepo.CheckOTP(email, otp)
	if err != nil {
		return errors.New("kode OTP salah")
	}

	// Cek Expiry
	if time.Now().After(reset.ExpiredAt) {
		u.authRepo.DeleteOTP(email) // Hapus yang sudah basi
		return errors.New("kode OTP sudah kadaluwarsa (lebih dari 15 menit)")
	}

	// Hash Password Baru
	// hashedPassword, _ := utils.HashPassword(newPassword)
	hashedPassword, err := utils.HashPassword(newPassword)
    if err != nil {
        return errors.New("gagal memproses password")
    }

	// Update di tabel Users
	// user, _ := u.userRepo.GetByEmail(email)
	// user.Password = hashedPassword
	// if err := u.userRepo.Update(&user); err != nil {
	// 	return err
	// }
	if err := u.userRepo.UpdatePasswordByEmail(email, hashedPassword); err != nil {
        return errors.New("gagal memperbarui password di database")
    }

	// Hapus OTP karena sudah berhasil dipakai
	return u.authRepo.DeleteOTP(email)
}

// Helper internal untuk generate angka
func (u *authUC) generateRandomOTP(max int) string {
	var table = []byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	b := make([]byte, max)
	n, _ := io.ReadAtLeast(rand.Reader, b, max)
	if n != max {
		return "123456" // Fallback jika crypto/rand gagal
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}
