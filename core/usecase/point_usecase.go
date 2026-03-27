package usecase

import (
	"ybg-backend-go/core/entity"
	"ybg-backend-go/core/repository"

	"github.com/google/uuid"
)

type PointUsecase interface {
	AddPointTransaction(uid uuid.UUID, point int) error
	GetMyPointHistory(uid uuid.UUID) ([]entity.PointHistory, error)
	FetchAllUsersPoints() ([]entity.PointTotal, error)
}

type pointUC struct {
	repo repository.PointRepository
}

func NewPointUsecase(repo repository.PointRepository) PointUsecase { return &pointUC{repo: repo} }

func (u *pointUC) AddPointTransaction(uid uuid.UUID, point int) error {
	// 1. Simpan ke history (PointID auto-increment)
	history := entity.PointHistory{
		UserID: uid,
		Point:  point,
	}
	if err := u.repo.CreateHistory(&history); err != nil {
		return err
	}

	// 2. Update Total di point_total
	return u.repo.UpdateTotal(uid, point)
}

func (u *pointUC) GetMyPointHistory(uid uuid.UUID) ([]entity.PointHistory, error) {
	return u.repo.GetHistoryByUserID(uid)
}

func (u *pointUC) FetchAllUsersPoints() ([]entity.PointTotal, error) {
	return u.repo.GetAllTotalsWithUser()
}
