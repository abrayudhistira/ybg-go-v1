package handler

import (
	"net/http"
	"os"

	"ybg-backend-go/core/delivery/http/middleware"
	"ybg-backend-go/core/repository"
	"ybg-backend-go/core/wire" // Pastikan import path ini benar

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var router *gin.Engine

func init() {
	_ = godotenv.Load()

	dsn := os.Getenv("DB_URL")
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		PrepareStmt: false,
	})
	if err != nil {
		return
	}

	// Seed Admin jika diperlukan
	repository.SeedAdmin(db)

	// Panggil Injector dari Wire
	userHandler := wire.InitializeUserHandler(db)
	productHandler := wire.InitializeProductHandler(db)
	newsHandler := wire.InitializeNewsHandler(db)
	brandHandler := wire.InitializeBrandHandler(db)
	categoryHandler := wire.InitializeCategoryHandler(db)
	pHandler := wire.InitializePointHandler(db)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// --- Routes Setup ---
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/health")
	})
	r.POST("/register", userHandler.Create)
	r.POST("/login", userHandler.Login)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "UP", "database": "connected"})
	})

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		// Groups & Handlers
		brandAdmin := api.Group("/brand")
		api.GET("/brand", brandHandler.GetAll)
		brandAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			brandAdmin.POST("/", brandHandler.Create)
			brandAdmin.DELETE("/:id", brandHandler.Delete)
		}

		categoryAdmin := api.Group("/category")
		api.GET("/category", categoryHandler.GetAll)
		categoryAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			categoryAdmin.POST("/", categoryHandler.Create)
			categoryAdmin.DELETE("/:id", categoryHandler.Delete)
		}

		api.GET("/products", productHandler.GetAll)
		api.GET("/products/:id", productHandler.GetByID)
		api.GET("/products/search", productHandler.Search)

		productAdmin := api.Group("/products")
		productAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			productAdmin.POST("/", productHandler.Create)
			productAdmin.PUT("/:id", productHandler.Update)
			productAdmin.DELETE("/:id", productHandler.Delete)
		}

		points := api.Group("/points")
		{
			points.GET("/history", pHandler.GetHistory)
			points.POST("/", middleware.RoleMiddleware("admin"), pHandler.CreatePoint)
			points.GET("/all", middleware.RoleMiddleware("admin"), pHandler.GetAllSummaries)
		}

		api.GET("/news", newsHandler.GetAll)
		newsAdmin := api.Group("/news")
		newsAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			newsAdmin.POST("/", newsHandler.Create)
			newsAdmin.PUT("/:id", newsHandler.Update)
			newsAdmin.DELETE("/:id", newsHandler.Delete)
		}

		api.GET("/users", middleware.RoleMiddleware("admin"), userHandler.GetAll)
		api.GET("/profile/:id", userHandler.GetByID)
		api.PUT("/profile/:id", userHandler.Update)
	}

	router = r
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if router == nil {
		http.Error(w, "Router not initialized", http.StatusInternalServerError)
		return
	}
	router.ServeHTTP(w, r)
}
