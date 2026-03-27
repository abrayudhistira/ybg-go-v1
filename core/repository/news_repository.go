package repository

import (
	"ybg-backend-go/core/entity"

	"gorm.io/gorm"
)

type NewsRepository interface {
	Create(n *entity.News) error
	GetAll() ([]entity.News, error)
	GetByID(id uint) (entity.News, error)
	Update(n *entity.News) error
	Delete(id uint) error
}

type newsRepo struct {
	db *gorm.DB
}

func NewNewsRepository(db *gorm.DB) NewsRepository {
	return &newsRepo{db: db}
}

func (r *newsRepo) Create(n *entity.News) error { return r.db.Create(n).Error }
func (r *newsRepo) GetAll() ([]entity.News, error) {
	var news []entity.News
	err := r.db.Order("created_at desc").Find(&news).Error
	return news, err
}
func (r *newsRepo) GetByID(id uint) (entity.News, error) {
	var n entity.News
	err := r.db.First(&n, id).Error
	return n, err
}
func (r *newsRepo) Update(n *entity.News) error { return r.db.Save(n).Error }
func (r *newsRepo) Delete(id uint) error        { return r.db.Delete(&entity.News{}, id).Error }
