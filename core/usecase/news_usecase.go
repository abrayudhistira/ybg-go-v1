package usecase

import (
	"ybg-backend-go/core/entity"
	"ybg-backend-go/core/repository"
)

type NewsUsecase interface {
	CreateNews(n *entity.News) error
	GetAllNews() ([]entity.News, error)
	GetNewsByID(id uint) (entity.News, error)
	UpdateNews(n *entity.News) error
	DeleteNews(id uint) error
}

type newsUC struct {
	repo repository.NewsRepository
}

func NewNewsUsecase(repo repository.NewsRepository) NewsUsecase {
	return &newsUC{repo: repo}
}

// ... Implementasi fungsi (panggil repo saja) ...
func (u *newsUC) CreateNews(n *entity.News) error          { return u.repo.Create(n) }
func (u *newsUC) GetAllNews() ([]entity.News, error)       { return u.repo.GetAll() }
func (u *newsUC) GetNewsByID(id uint) (entity.News, error) { return u.repo.GetByID(id) }
func (u *newsUC) UpdateNews(n *entity.News) error          { return u.repo.Update(n) }
func (u *newsUC) DeleteNews(id uint) error                 { return u.repo.Delete(id) }
