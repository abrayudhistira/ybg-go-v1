package repository

import (
	"log"
	"time"
	"ybg-backend-go/core/entity"
	"ybg-backend-go/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedAdmin(db *gorm.DB) {
	birthDate, err := time.Parse("02-01-2006", "10-10-2003")
	if err != nil {
		log.Fatalf("Invalid birth date format: %v", err)
	}
	adminID := uuid.New() // or use uuid.New().String() if you want a unique ID
	password, _ := utils.HashPassword("ybsupabase@")
	admin := entity.User{
		UserID:   adminID,
		Name:     "Admin YBG",
		Email:    "ybsupabase@gmail.com",
		Birth:    birthDate,
		Password: password,
		Role:     "admin",
		Phone:    "087840866596",
		Gender:   "male",
	}

	// Cek apakah admin sudah ada berdasarkan email
	var exists int64
	db.Model(&entity.User{}).Where("email = ?", admin.Email).Count(&exists)

	if exists == 0 {
		err := db.Transaction(func(tx *gorm.DB) error {
			// Simpan User
			if err := tx.Create(&admin).Error; err != nil {
				return err
			}
			// Simpan Point Total awal
			point := entity.PointTotal{
				UserID: adminID,
				Total:  0,
				Tier:   "friend",
			}
			return tx.Create(&point).Error
		})

		if err != nil {
			log.Printf("Gagal seeding admin: %v", err)
		} else {
			log.Println("✅ Admin user seeded successfully!")
		}
	}
}
