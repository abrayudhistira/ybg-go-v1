package entity

import (
	"time"

	"github.com/google/uuid"
)

// --- USER & LOYALTY ENTITIES ---

type User struct {
	UserID uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	Name   string    `gorm:"size:100;not null" json:"name" binding:"required,min=3"`
	Email  string    `gorm:"size:100;unique;not null" json:"email" binding:"required,email"`
	// Password       string         `gorm:"size:255;not null" json:"-"`
	Password       string         `gorm:"size:255;not null" json:"password"`
	ProfilePicture string         `json:"profile_picture"`
	Birth          time.Time      `gorm:"not null" json:"birth" binding:"required"`
	Role           string         `gorm:"type:user_role;not null;default:customer" json:"role"`
	Phone          string         `gorm:"size:13;not null" json:"phone" binding:"required"`
	Gender         string         `gorm:"type:user_gender;not null" json:"gender" binding:"required"`
	IsVerified     bool           `gorm:"default:false;not null" json:"is_verified"`
	CreatedAt      time.Time      `gorm:"not null;default:now()" json:"created_at"`
	PointTotal     *PointTotal    `gorm:"foreignKey:UserID" json:"point_total"`
	PointHistory   []PointHistory `gorm:"foreignKey:UserID" json:"point_history"`
}

func (User) TableName() string { return "users" }

type PointTotal struct {
	TotalPointID uint      `gorm:"primaryKey" json:"total_point_id"`
	UserID       uuid.UUID `gorm:"type:uuid;uniqueIndex" json:"user_id"`
	Total        int       `gorm:"default:0" json:"total"`
	Tier         string    `gorm:"type:point_tier;default:friend" json:"tier"`
	CreatedAt    time.Time `json:"created_at"`
	User         *User     `gorm:"foreignKey:UserID;references:UserID" json:"user"`
}

func (PointTotal) TableName() string { return "point_total" }

type PointHistory struct {
	PointID   uint       `gorm:"primaryKey;column:point_id" json:"point_id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	Point     int        `gorm:"column:point;not null" json:"point"`
	Status    string     `gorm:"type:varchar;default:aktif" json:"status"`
	ExpiredAt *time.Time `gorm:"column:expired_at" json:"expired_at"`
	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
}

func (PointHistory) TableName() string { return "point_history" }

// --- PRODUCT CATALOG ENTITIES ---

type Brand struct {
	BrandID   uint      `gorm:"primaryKey" json:"brand_id"`
	Name      string    `gorm:"size:50;not null" json:"name" form:"name"`
	ImageURL  string    `gorm:"column:image_url" json:"image_url"`
	CreatedAt time.Time `json:"created_at"`
	Products  []Product `gorm:"foreignKey:BrandID" json:"products,omitempty"`
}

func (Brand) TableName() string { return "brand" }

type Category struct {
	CategoryID uint      `gorm:"primaryKey" json:"category_id"`
	Name       string    `gorm:"size:50;not null" json:"name"`
	CreatedAt  time.Time `json:"created_at"`
	Products   []Product `gorm:"foreignKey:CategoryID" json:"products,omitempty"`
}

func (Category) TableName() string { return "category" }

type Product struct {
	ProductID   uint      `gorm:"primaryKey" json:"product_id"`
	BrandID     uint      `json:"brand_id"`
	CategoryID  uint      `json:"category_id"`
	Name        string    `gorm:"size:50;not null" json:"name"`
	Price       float64   `gorm:"type:numeric" json:"price"`
	ImageURL    string    `json:"image_url"`
	Description string    `json:"description"`
	Stock       int       `json:"stock"`
	Condition   string    `gorm:"type:product_condition" json:"condition"`
	Status      string    `gorm:"type:product_status" json:"status"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedAt   time.Time `json:"created_at"`
	Brand       Brand     `gorm:"foreignKey:BrandID" json:"brand"`
	Category    Category  `gorm:"foreignKey:CategoryID" json:"category"`
}

func (Product) TableName() string { return "product" } // Memaksa ke tabel 'product' bukan 'products'

// --- INFORMATION ENTITY ---

type News struct {
	NewsID      uint      `gorm:"primaryKey" json:"news_id"`
	Title       string    `gorm:"size:50;not null" json:"title"`
	Description string    `json:"description"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	Status      string    `gorm:"type:news_status" json:"status"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedAt   time.Time `json:"created_at"`
}

func (News) TableName() string { return "news" }

type PasswordReset struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"size:100;not null" json:"email"`
	OTP       string    `gorm:"size:6;not null" json:"otp"`
	ExpiredAt time.Time `gorm:"not null" json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (PasswordReset) TableName() string { return "password_resets" }

type Cart struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;unique;not null" json:"user_id"` // Gunakan uuid.UUID
	Items     []CartItem `gorm:"foreignKey:CartID" json:"items"`
	CreatedAt time.Time  `json:"created_at"`
}

func (Cart) TableName() string { return "carts" }

// CartItem adalah isi dari keranjang yang merujuk ke Product
type CartItem struct {
	ID        uint `gorm:"primaryKey" json:"id"`
	CartID    uint `gorm:"column:cart_id" json:"cart_id"`       // Eksplisit kolom database
	ProductID uint `gorm:"column:product_id" json:"product_id"` // Eksplisit kolom database

	// Tambahkan references:ProductID karena PK di struct Product bukan bernama 'ID'
	Product Product `gorm:"foreignKey:ProductID;references:ProductID" json:"product"`

	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (CartItem) TableName() string { return "cart_items" }

type Reward struct {
	RewardID    uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"reward_id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Description string    `json:"description"`
	PointCost   int       `gorm:"not null" json:"point_cost"`
	Quantity    int       `gorm:"not null" json:"quantity"`
	Category    string    `gorm:"type:reward_category;not null;default:voucher" json:"category"`
	ImageURL    string    `json:"image_url"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:now();autoUpdateTime" json:"updated_at"`
}

func (Reward) TableName() string { return "rewards" }

type RewardHistory struct {
	HistoryID  uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"history_id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	RewardID   uuid.UUID `gorm:"type:uuid;not null" json:"reward_id"`
	PointSpent int       `gorm:"not null" json:"point_spent"`
	Status     string    `gorm:"type:reward_history_status;not null;default:pengajuan" json:"status"`
	AdminNote  string    `json:"admin_note"`
	CreatedAt  time.Time `gorm:"column:created_at;not null;default:now()" json:"created_at"`

	// Relasi (Belongs To)
	Reward Reward `gorm:"foreignKey:RewardID;references:RewardID" json:"reward"`
	User   User   `gorm:"foreignKey:UserID;references:UserID" json:"user"`
}

func (RewardHistory) TableName() string { return "reward_histories" }
