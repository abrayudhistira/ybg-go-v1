package repository

import (
	"ybg-backend-go/core/entity"

	"gorm.io/gorm"
)

type AuthRepository interface {
	SaveOTP(reset *entity.PasswordReset) error
	CheckOTP(email, otp string) (*entity.PasswordReset, error)
	DeleteOTP(email string) error
}

type authRepo struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepo{db: db}
}

func (r *authRepo) SaveOTP(reset *entity.PasswordReset) error {
	// Hapus OTP lama jika ada sebelum simpan yang baru (biar tidak numpuk)
	r.db.Where("email = ?", reset.Email).Delete(&entity.PasswordReset{})
	return r.db.Create(reset).Error
}

func (r *authRepo) CheckOTP(email, otp string) (*entity.PasswordReset, error) {
	var reset entity.PasswordReset
	err := r.db.Where("email = ? AND otp = ?", email, otp).First(&reset).Error
	return &reset, err
}

func (r *authRepo) DeleteOTP(email string) error {
	return r.db.Where("email = ?", email).Delete(&entity.PasswordReset{}).Error
}
