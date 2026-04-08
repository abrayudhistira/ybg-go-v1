//go:build wireinject
// +build wireinject

package wire

import (
	"ybg-backend-go/core/delivery/http"
	"ybg-backend-go/core/repository"
	"ybg-backend-go/core/usecase"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// Definisikan semua Set Dependency
var userSet = wire.NewSet(
	repository.NewUserRepository,
	repository.NewPointRepository,
	usecase.NewUserUsecase,
	http.NewUserHandler,
)

var productSet = wire.NewSet(
	repository.NewProductRepository,
	usecase.NewProductUsecase,
	http.NewProductHandler,
)

var newsSet = wire.NewSet(
	repository.NewNewsRepository,
	usecase.NewNewsUsecase,
	http.NewNewsHandler,
)

var brandSet = wire.NewSet(
	repository.NewBrandRepository,
	usecase.NewBrandUsecase,
	http.NewBrandHandler,
)

var categorySet = wire.NewSet(
	repository.NewCategoryRepository,
	usecase.NewCategoryUsecase,
	http.NewCategoryHandler,
)

var pointSet = wire.NewSet(
	repository.NewPointRepository,
	usecase.NewPointUsecase,
	http.NewPointHandler,
)

var cartSet = wire.NewSet(
	repository.NewCartRepository,
	repository.NewProductRepository, // Dibutuhkan untuk cek stok di usecase cart
	usecase.NewCartUsecase,
	http.NewCartHandler,
)

var authSet = wire.NewSet(
	repository.NewAuthRepository, // Pastikan ini sudah dibuat di folder repository
	repository.NewUserRepository, // Dibutuhkan untuk update password
	usecase.NewAuthUsecase,
	http.NewAuthHandler, // Pastikan nama package handler kamu adalah 'http' sesuai yang lain
)

var rewardSet = wire.NewSet(
	repository.NewRewardRepository,
	repository.NewPointRepository, // Dibutuhkan oleh usecase reward
	usecase.NewRewardUsecase,
	http.NewRewardHandler,
)

func InitializeAuthHandler(db *gorm.DB) *http.AuthHandler {
	wire.Build(authSet)
	return nil
}

// Injector Functions
func InitializeUserHandler(db *gorm.DB) *http.UserHandler {
	wire.Build(userSet)
	return nil
}

func InitializeProductHandler(db *gorm.DB) *http.ProductHandler {
	wire.Build(productSet)
	return nil
}

func InitializeNewsHandler(db *gorm.DB) *http.NewsHandler {
	wire.Build(newsSet)
	return nil
}

func InitializeBrandHandler(db *gorm.DB) *http.BrandHandler {
	wire.Build(brandSet)
	return nil
}

func InitializeCategoryHandler(db *gorm.DB) *http.CategoryHandler {
	wire.Build(categorySet)
	return nil
}

func InitializePointHandler(db *gorm.DB) *http.PointHandler {
	wire.Build(pointSet)
	return nil
}
func InitializeCartHandler(db *gorm.DB) *http.CartHandler {
	wire.Build(cartSet)
	return nil
}
func InitializeRewardHandler(db *gorm.DB) *http.RewardHandler {
	wire.Build(rewardSet)
	return nil
}
