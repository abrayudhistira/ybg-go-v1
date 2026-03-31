package usecase

import (
	"errors"
	"ybg-backend-go/core/entity"
	"ybg-backend-go/core/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CartUsecase interface {
	GetCart(userID string) (*entity.Cart, error)
	AddItemToCart(userID string, productID uint, qty int) error
	RemoveItemFromCart(userID string, productID uint) error
	ClearCart(userID string) error
}

type cartUC struct {
	repo        repository.CartRepository
	productRepo repository.ProductRepository // Kita butuh ini buat cek stok
}

func NewCartUsecase(r repository.CartRepository, p repository.ProductRepository) CartUsecase {
	return &cartUC{repo: r, productRepo: p}
}

func (u *cartUC) GetCart(userID string) (*entity.Cart, error) {
	return u.repo.GetByUserID(userID)
}

func (u *cartUC) AddItemToCart(userID string, productID uint, qty int) error {
	// 1. Parse string userID (dari middleware) ke tipe uuid.UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("format user id tidak valid")
	}

	// 2. Validasi Produk & Stok
	// Pastikan productRepo.GetByID mengembalikan detail product (termasuk stok)
	product, err := u.productRepo.GetByID(productID)
	if err != nil {
		return errors.New("produk tidak ditemukan")
	}

	if product.Stock < qty {
		return errors.New("stok produk tidak mencukupi")
	}

	// 3. Cari Keranjang (Cart) milik user tersebut
	// Kita tetap kirim string userID ke repo jika interface repo-nya minta string
	cart, err := u.repo.GetByUserID(userID)

	if err != nil {
		// Jika keranjang belum ada (Record Not Found), buat baru
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newCart := &entity.Cart{
				UserID: userUUID, // Menggunakan hasil parse UUID tadi
			}
			if err := u.repo.CreateCart(newCart); err != nil {
				return errors.New("gagal membuat keranjang baru")
			}
			// Set variabel cart ke newCart yang baru dibuat agar proses di bawah bisa lanjut
			cart = newCart
		} else {
			// Jika error lain (koneksi db, dll)
			return err
		}
	}

	// 4. Cek apakah produk ini SUDAH ADA di dalam keranjang (CartItem)
	// Kita cari berdasarkan CartID dan ProductID
	item, err := u.repo.GetItem(cart.ID, productID)

	if err == nil {
		// KONDISI: Produk sudah ada, tinggal update quantity
		// Tambahkan validasi lagi: total quantity baru tidak boleh melebihi stok
		if product.Stock < (item.Quantity + qty) {
			return errors.New("total pesanan melebihi stok yang tersedia")
		}

		item.Quantity += qty
		return u.repo.UpdateItem(item)
	}

	// 5. KONDISI: Produk belum ada di keranjang, buat row baru di cart_items
	newItem := &entity.CartItem{
		CartID:    cart.ID,
		ProductID: productID,
		Quantity:  qty,
	}

	return u.repo.AddItem(newItem)
}

func (u *cartUC) RemoveItemFromCart(userID string, productID uint) error {
	cart, err := u.repo.GetByUserID(userID)
	if err != nil {
		return errors.New("keranjang tidak ditemukan")
	}
	return u.repo.DeleteItem(cart.ID, productID)
}
func (u *cartUC) ClearCart(userID string) error {
	cart, err := u.repo.GetByUserID(userID)
	if err != nil {
		return errors.New("keranjang tidak ditemukan")
	}

	return u.repo.DeleteAllItem(cart.ID)
}
