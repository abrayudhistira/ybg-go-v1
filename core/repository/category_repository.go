package repository

import (
	"ybg-backend-go/core/entity"

	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(c *entity.Category) error
	GetAll() ([]entity.Category, error)
	Delete(id uint) error
}

type categoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository { return &categoryRepo{db: db} }

func (r *categoryRepo) Create(c *entity.Category) error { return r.db.Create(c).Error }
func (r *categoryRepo) GetAll() ([]entity.Category, error) {
	var categories []entity.Category
	err := r.db.Find(&categories).Error
	return categories, err
}
func (r *categoryRepo) Delete(id uint) error { return r.db.Delete(&entity.Category{}, id).Error }
