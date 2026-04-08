package handler

import (
	"log"
	"net/http"
	"os"
	"time"

	"ybg-backend-go/core/delivery/http/middleware"
	"ybg-backend-go/core/repository"
	"ybg-backend-go/core/wire"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var router *gin.Engine

func init() {
	// 1. Load ENV
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found, using system environment variables")
	}

	// 2. Database Connection
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		log.Println("CRITICAL ERROR: DB_URL is empty!")
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		PrepareStmt: false,
	})

	if err != nil {
		log.Printf("FATAL DATABASE ERROR: %v\n", err)
		return
	}
	log.Println("SUCCESS: Database connected")

	// 3. Panic Recovery for Initialization
	// Jika Wire gagal karena ketidakcocokan parameter, log akan muncul di sini
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC RECOVERED during Wire/Routes init: %v\n", r)
		}
	}()

	// 4. Seeding & Dependency Injection
	repository.SeedAdmin(db)

	userHandler := wire.InitializeUserHandler(db)
	productHandler := wire.InitializeProductHandler(db)
	newsHandler := wire.InitializeNewsHandler(db)
	brandHandler := wire.InitializeBrandHandler(db)
	categoryHandler := wire.InitializeCategoryHandler(db)
	pHandler := wire.InitializePointHandler(db)
	authHandler := wire.InitializeAuthHandler(db)
	cartHandler := wire.InitializeCartHandler(db)
	rewardHandler := wire.InitializeRewardHandler(db)

	// 5. Gin Setup
	gin.SetMode(gin.ReleaseMode)
	r := gin.New() // Menggunakan New() untuk kontrol middleware penuh

	// Middleware bawaan untuk logging setiap request ke terminal
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

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
		// Brand
		brandAdmin := api.Group("/brand")
		api.GET("/brand", brandHandler.GetAll)
		brandAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			brandAdmin.POST("/admin", brandHandler.Create)
			brandAdmin.DELETE("/admin/:id", brandHandler.Delete)
			brandAdmin.PUT("/admin/:id", brandHandler.Update)
		}

		// Category
		categoryAdmin := api.Group("/category")
		api.GET("/category", categoryHandler.GetAll)
		categoryAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			categoryAdmin.POST("/admin", categoryHandler.Create)
			categoryAdmin.DELETE("/admin/:id", categoryHandler.Delete)
		}

		// Product
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

		// Points
		points := api.Group("/points")
		{
			points.GET("/history", pHandler.GetHistory)
			points.POST("/admin", middleware.RoleMiddleware("admin"), pHandler.CreatePoint)
			points.GET("/admin/all", middleware.RoleMiddleware("admin"), pHandler.GetAllSummaries)
		}

		// News
		api.GET("/news", newsHandler.GetAll)
		newsAdmin := api.Group("/news")
		newsAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			newsAdmin.POST("/admin", newsHandler.Create)
			newsAdmin.PUT("/admin/:id", newsHandler.Update)
			newsAdmin.DELETE("/admin/:id", newsHandler.Delete)
		}

		api.GET("/admin/users", middleware.RoleMiddleware("admin"), userHandler.GetAll)

		// Profile
		userGroup := api.Group("/profile")
		{
			userGroup.GET("/:id", userHandler.GetByID)
			userGroup.PUT("/:id", userHandler.Update)
			userGroup.POST("/request-change-email", authHandler.RequestChangeEmail)
			userGroup.POST("/verify-change-email", authHandler.VerifyChangeEmail)
		}

		// Cart
		cartGroup := api.Group("/cart")
		{
			cartGroup.GET("/", cartHandler.GetMyCart)
			cartGroup.POST("/", cartHandler.AddToCart)
			cartGroup.DELETE("/:id", cartHandler.DeleteItem)
			cartGroup.DELETE("/clear", cartHandler.ClearMyCart)
		}

		// Rewards User
		rewards := api.Group("/rewards")
		{
			rewards.GET("/", rewardHandler.GetAll)
			rewards.POST("/claim", rewardHandler.Claim)
			rewards.GET("/history", rewardHandler.GetMyHistory)
		}

		// Rewards Admin
		rewardsAdmin := api.Group("/rewards/admin")
		rewardsAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			rewardsAdmin.PATCH("/approve", rewardHandler.Approve)
			rewardsAdmin.PATCH("/reject", rewardHandler.Reject)
			rewardsAdmin.POST("/", rewardHandler.Create)
			rewardsAdmin.PUT("/:id", rewardHandler.Update)
			rewardsAdmin.DELETE("/:id", rewardHandler.Delete)
		}
	}

	router = r
	log.Println("SUCCESS: Router fully initialized")
}

// Handler is the entry point for Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	// Logging request info ke terminal Vercel/Local
	log.Printf("INCOMING REQUEST: %s %s", r.Method, r.URL.Path)

	if router == nil {
		log.Println("ERROR: Router is NIL. This is likely due to a DB failure or init() panic.")
		http.Error(w, "Internal Server Error: Router not initialized. Check logs.", http.StatusInternalServerError)
		return
	}
	router.ServeHTTP(w, r)
}
