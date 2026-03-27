package usecase

import (
	"ybg-backend-go/core/entity"
	"ybg-backend-go/core/repository"
)

type BrandUsecase interface {
	CreateBrand(b *entity.Brand) error
	GetAllBrands() ([]entity.Brand, error)
	DeleteBrand(id uint) error
}

type brandUC struct {
	repo repository.BrandRepository
}

func NewBrandUsecase(repo repository.BrandRepository) BrandUsecase { return &brandUC{repo: repo} }

func (u *brandUC) CreateBrand(b *entity.Brand) error     { return u.repo.Create(b) }
func (u *brandUC) GetAllBrands() ([]entity.Brand, error) { return u.repo.GetAll() }
func (u *brandUC) DeleteBrand(id uint) error             { return u.repo.Delete(id) }
