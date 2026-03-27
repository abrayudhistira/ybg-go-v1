package repository

import (
	"ybg-backend-go/core/entity"

	"gorm.io/gorm"
)

type BrandRepository interface {
	Create(b *entity.Brand) error
	GetAll() ([]entity.Brand, error)
	Delete(id uint) error
}

type brandRepo struct {
	db *gorm.DB
}

func NewBrandRepository(db *gorm.DB) BrandRepository { return &brandRepo{db: db} }

func (r *brandRepo) Create(b *entity.Brand) error { return r.db.Create(b).Error }
func (r *brandRepo) GetAll() ([]entity.Brand, error) {
	var brands []entity.Brand
	err := r.db.Find(&brands).Error
	return brands, err
}
func (r *brandRepo) Delete(id uint) error { return r.db.Delete(&entity.Brand{}, id).Error }
