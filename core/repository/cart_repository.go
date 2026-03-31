package repository

import (
	"errors"
	"ybg-backend-go/core/entity"

	"gorm.io/gorm"
)

type CartRepository interface {
	GetByUserID(userID string) (*entity.Cart, error)
	CreateCart(cart *entity.Cart) error
	GetItem(cartID uint, productID uint) (*entity.CartItem, error)
	AddItem(item *entity.CartItem) error
	UpdateItem(item *entity.CartItem) error
	DeleteItem(cartID uint, productID uint) error
	DeleteAllItem(cartID uint) error
}

type cartRepo struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) CartRepository {
	return &cartRepo{db: db}
}

func (r *cartRepo) GetByUserID(userID string) (*entity.Cart, error) {
	var cart entity.Cart
	// Nested Preload: Cart -> Items -> Product -> Brand & Category
	err := r.db.Preload("Items.Product.Brand").
		Preload("Items.Product.Category").
		Where("user_id = ?", userID).
		First(&cart).Error
	return &cart, err
}

func (r *cartRepo) CreateCart(cart *entity.Cart) error {
	return r.db.Create(cart).Error
}

func (r *cartRepo) GetItem(cartID uint, productID uint) (*entity.CartItem, error) {
	var item entity.CartItem
	err := r.db.Where("cart_id = ? AND product_id = ?", cartID, productID).First(&item).Error
	return &item, err
}

func (r *cartRepo) AddItem(item *entity.CartItem) error {
	return r.db.Create(item).Error
}

func (r *cartRepo) UpdateItem(item *entity.CartItem) error {
	return r.db.Save(item).Error
}

func (r *cartRepo) DeleteItem(cartID uint, productID uint) error {
	// Gunakan .Debug() supaya kamu bisa lihat query aslinya di terminal
	result := r.db.Debug().Where("cart_id = ? AND product_id = ?", cartID, productID).Delete(&entity.CartItem{})

	if result.Error != nil {
		return result.Error
	}

	// Cek apakah ada baris yang benar-benar terhapus
	if result.RowsAffected == 0 {
		return errors.New("data tidak ditemukan, gagal menghapus")
	}

	return nil
}
func (r *cartRepo) DeleteAllItem(cartID uint) error {
	// Menghapus SEMUA row di tabel cart_items yang punya cart_id tersebut
	return r.db.Debug().Where("cart_id = ?", cartID).Delete(&entity.CartItem{}).Error
}
