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
