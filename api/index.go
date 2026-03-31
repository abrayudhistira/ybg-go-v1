package handler

import (
	"net/http"
	"os"
	"time"

	"ybg-backend-go/core/delivery/http/middleware"
	"ybg-backend-go/core/repository"
	"ybg-backend-go/core/wire" // Pastikan import path ini benar

	"github.com/gin-contrib/cors"
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
	authHandler := wire.InitializeAuthHandler(db)
	cartHandler := wire.InitializeCartHandler(db)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// --- Routes Setup ---
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/health")
	})
	r.POST("/register", userHandler.Create)
	r.POST("/login", userHandler.Login)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "UP", "database": "connected"})
	})

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/forgot-password", authHandler.ForgotPassword)
		authGroup.POST("/reset-password", authHandler.ResetPassword)
	}

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		// Groups & Handlers
		brandAdmin := api.Group("/brand")
		api.GET("/brand", brandHandler.GetAll)
		brandAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			brandAdmin.POST("/admin", brandHandler.Create)
			brandAdmin.DELETE("/admin/:id", brandHandler.Delete)
			brandAdmin.PUT("/admin/:id", brandHandler.Update)
		}

		categoryAdmin := api.Group("/category")
		api.GET("/category", categoryHandler.GetAll)
		categoryAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			categoryAdmin.POST("/admin", categoryHandler.Create)
			categoryAdmin.DELETE("/admin/:id", categoryHandler.Delete)
		}

		api.GET("/products", productHandler.GetAll)
		api.GET("/products/:id", productHandler.GetByID)
		api.GET("/products/search", productHandler.Search)

		productAdmin := api.Group("/products")
		productAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			productAdmin.POST("/admin", productHandler.Create)
			productAdmin.PUT("/admin/:id", productHandler.Update)
			productAdmin.DELETE("/admin/:id", productHandler.Delete)
		}

		points := api.Group("/points")
		{
			points.GET("/history", pHandler.GetHistory)
			points.POST("/admin", middleware.RoleMiddleware("admin"), pHandler.CreatePoint)
			points.GET("/admin/all", middleware.RoleMiddleware("admin"), pHandler.GetAllSummaries)
		}

		api.GET("/news", newsHandler.GetAll)
		newsAdmin := api.Group("/news")
		newsAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			newsAdmin.POST("/admin", newsHandler.Create)
			newsAdmin.PUT("/admin/:id", newsHandler.Update)
			newsAdmin.DELETE("/admin/:id", newsHandler.Delete)
		}

		api.GET("/admin/users", middleware.RoleMiddleware("admin"), userHandler.GetAll)
		api.GET("/profile/:id", userHandler.GetByID)
		api.PUT("/profile/:id", userHandler.Update)

		cartGroup := api.Group("/cart")
		{
			cartGroup.GET("/", cartHandler.GetMyCart)
			cartGroup.POST("/", cartHandler.AddToCart)
			cartGroup.DELETE("/:id", cartHandler.DeleteItem)
			cartGroup.DELETE("/clear", cartHandler.ClearMyCart)
		}
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
