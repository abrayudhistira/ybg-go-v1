package repository

import (
	"ybg-backend-go/core/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RewardRepository interface {
	GetAll() ([]entity.Reward, error)
	GetByID(id uuid.UUID) (entity.Reward, error)
	CreateHistory(history *entity.RewardHistory) error
	GetHistoryByUserID(userID uuid.UUID) ([]entity.RewardHistory, error)
	UpdateQuantity(id uuid.UUID, newQty int) error
	GetHistoryByID(id uuid.UUID) (entity.RewardHistory, error)
	UpdateHistoryStatus(id uuid.UUID, status string) error
	Create(reward *entity.Reward) error
}

type rewardRepo struct {
	db *gorm.DB
}

func NewRewardRepository(db *gorm.DB) RewardRepository {
	return &rewardRepo{db: db}
}

func (r *rewardRepo) GetAll() ([]entity.Reward, error) {
	var rewards []entity.Reward
	err := r.db.Find(&rewards).Error
	return rewards, err
}

func (r *rewardRepo) GetByID(id uuid.UUID) (entity.Reward, error) {
	var reward entity.Reward
	err := r.db.First(&reward, "reward_id = ?", id).Error
	return reward, err
}

func (r *rewardRepo) CreateHistory(history *entity.RewardHistory) error {
	return r.db.Create(history).Error
}

func (r *rewardRepo) GetHistoryByUserID(userID uuid.UUID) ([]entity.RewardHistory, error) {
	var histories []entity.RewardHistory
	// Kita Preload("Reward") supaya data detail reward-nya ikut terbawa
	err := r.db.Preload("Reward").Where("user_id = ?", userID).Find(&histories).Error
	return histories, err
}

func (r *rewardRepo) UpdateQuantity(id uuid.UUID, newQty int) error {
	return r.db.Model(&entity.Reward{}).Where("reward_id = ?", id).Update("quantity", newQty).Error
}
func (r *rewardRepo) GetHistoryByID(id uuid.UUID) (entity.RewardHistory, error) {
	var history entity.RewardHistory
	// Kita pakai Preload("Reward") supaya di usecase bisa dapet detail reward-nya (kayak PointCost)
	err := r.db.Preload("Reward").First(&history, "history_id = ?", id).Error
	return history, err
}

func (r *rewardRepo) UpdateHistoryStatus(id uuid.UUID, status string) error {
	return r.db.Model(&entity.RewardHistory{}).
		Where("history_id = ?", id).
		Update("status", status).Error
}

func (r *rewardRepo) Create(reward *entity.Reward) error {
	return r.db.Create(reward).Error
}
