package usecase

import (
	"crypto/rand"
	"errors"
	"io"
	"time"
	"ybg-backend-go/core/entity"
	"ybg-backend-go/core/repository"
	"ybg-backend-go/pkg/utils"

	"github.com/google/uuid"
)

type AuthUsecase interface {
	RequestOTP(email string) error
	VerifyOTPAndResetPassword(email, otp, newPassword string) error
	RequestEmailUpdateOTP(newEmail string) error
	VerifyAndChangeEmail(userID uuid.UUID, email, otp string) error
	VerifyAccount(email, otp string) error
	ResendVerificationOTP(email string) error
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
func (u *authUC) RequestEmailUpdateOTP(newEmail string) error {
	// 1. Validasi: Cek apakah email baru sudah ada yang punya?
	existingUser, _ := u.userRepo.GetByEmail(newEmail)
	if existingUser.Email != "" {
		return errors.New("email sudah terdaftar oleh pengguna lain")
	}

	// 2. Generate 6 digit angka random
	otp := u.generateRandomOTP(6)

	// 3. Simpan ke DB (Expired 15 menit)
	// Kita pakai entity PasswordReset yang sudah ada karena strukturnya sama (Email & OTP)
	resetData := &entity.PasswordReset{
		Email:     newEmail,
		OTP:       otp,
		ExpiredAt: time.Now().Add(15 * time.Minute),
	}

	if err := u.authRepo.SaveOTP(resetData); err != nil {
		return err
	}

	// 4. Kirim ke email BARU tersebut
	// User harus bisa buka email barunya untuk verifikasi
	return utils.SendOTPEmail(newEmail, otp)
}
func (u *authUC) VerifyAndChangeEmail(userID uuid.UUID, email, otp string) error {
	// 1. Cek OTP di AuthRepo (Tabel password_resets)
	reset, err := u.authRepo.CheckOTP(email, otp)
	if err != nil {
		return errors.New("kode OTP salah")
	}

	// 2. Cek apakah OTP sudah expired
	if time.Now().After(reset.ExpiredAt) {
		u.authRepo.DeleteOTP(email)
		return errors.New("kode OTP sudah kadaluwarsa")
	}

	// 3. EKSEKUSI: Update email di tabel Users via UserRepository
	if err := u.userRepo.UpdateEmail(userID, email); err != nil {
		return errors.New("gagal memperbarui email di database")
	}

	// 4. Bersihkan OTP karena sudah terpakai
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

// core/usecase/auth_usecase.go

func (u *authUC) VerifyAccount(email, otp string) error {
	// 1. Cek apakah OTP valid di tabel password_resets
	reset, err := u.authRepo.CheckOTP(email, otp)
	if err != nil {
		return errors.New("kode OTP salah atau tidak ditemukan")
	}

	// 2. Cek apakah OTP sudah kadaluwarsa
	if time.Now().After(reset.ExpiredAt) {
		u.authRepo.DeleteOTP(email) // Bersihkan yang expired
		return errors.New("kode OTP sudah kadaluwarsa, silakan minta kode baru")
	}

	// 3. Ambil data user berdasarkan email
	user, err := u.userRepo.GetByEmail(email)
	if err != nil {
		return errors.New("user tidak ditemukan")
	}

	// 4. Update status is_verified menjadi true
	// Kita gunakan repository user untuk mengupdate kolom is_verified
	if err := u.userRepo.VerifyUser(user.UserID); err != nil {
		return errors.New("gagal memverifikasi akun")
	}

	// 5. Hapus OTP dari database karena sudah berhasil digunakan
	return u.authRepo.DeleteOTP(email)
}

func (u *authUC) ResendVerificationOTP(email string) error {
	// 1. Pastikan user ada
	user, err := u.userRepo.GetByEmail(email)
	if err != nil {
		return errors.New("email tidak terdaftar")
	}

	// 2. Cek apakah sudah verified? Kalau sudah, tidak perlu kirim lagi
	if user.IsVerified {
		return errors.New("akun ini sudah terverifikasi, silakan langsung login")
	}

	// 3. Generate & Simpan OTP Baru (Gunakan helper generateRandomOTP yang sudah kamu buat)
	otp := u.generateRandomOTP(6)
	resetData := &entity.PasswordReset{
		Email:     email,
		OTP:       otp,
		ExpiredAt: time.Now().Add(15 * time.Minute),
	}

	if err := u.authRepo.SaveOTP(resetData); err != nil {
		return err
	}

	// 4. Kirim Email
	return utils.SendOTPEmail(email, otp)
}
