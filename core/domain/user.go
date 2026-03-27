package domain

import (
	"context"
	"time"
	"github.com/google/uuid"
)

type User struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"user_id"`
	FullName  string    `gorm:"type:varchar(255);not null" json:"full_name"`
	Gender    string    `gorm:"type:user_gender" json:"gender"` // user_gender adalah Enum di DB kamu
	Role      string    `gorm:"type:user_role;default:customer" json:"role"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// Interface tetap sama, ini kunci Clean Architecture
type UserUsecase interface {
	GetProfile(ctx context.Context, id string) (*User, error)
}

type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
}