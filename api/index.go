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
	r := gin.New()

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

	// --- GLOBAL ROUTES ---
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/health")
	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "UP", "database": "connected"})
	})

	// --- AUTH & USER REGISTRATION ---
	r.POST("/register", userHandler.Create)
	r.POST("/login", userHandler.Login)
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/forgot-password", authHandler.ForgotPassword)
		authGroup.POST("/reset-password", authHandler.ResetPassword)
		authGroup.POST("/verify-registration", authHandler.VerifyRegistration)
		authGroup.POST("/resend-otp", authHandler.ResendOTP)
	}

	// --- API PROTECTED AREA ---
	api := r.Group("/api")
	{
		// PUBLIC GET (No Auth Needed for Reading)
		api.GET("/products", productHandler.GetAll)
		api.GET("/products/:id", productHandler.GetByID)
		api.GET("/products/search", productHandler.Search)
		api.GET("/news", newsHandler.GetAll)
		api.GET("/news/:id", newsHandler.GetByID)
		api.GET("/brand", brandHandler.GetAll)
		api.GET("/category", categoryHandler.GetAll)

		// AUTHENTICATION MIDDLEWARE STARTS HERE
		api.Use(middleware.AuthMiddleware())

		// News Management (Admin Only)
		newsAdmin := api.Group("/news/admin")
		newsAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			newsAdmin.POST("/", newsHandler.Create)
			newsAdmin.PUT("/:id", newsHandler.Update)
			newsAdmin.DELETE("/:id", newsHandler.Delete)
		}

		// Brand Management (Admin Only)
		brandAdmin := api.Group("/brand/admin")
		brandAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			brandAdmin.POST("/", brandHandler.Create)
			brandAdmin.PUT("/:id", brandHandler.Update)
			brandAdmin.DELETE("/:id", brandHandler.Delete)
		}

		// Category Management (Admin Only)
		categoryAdmin := api.Group("/category/admin")
		categoryAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			categoryAdmin.POST("/", categoryHandler.Create)
			categoryAdmin.DELETE("/:id", categoryHandler.Delete)
		}

		// Product Management (Admin Only)
		productAdmin := api.Group("/products/admin")
		productAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			productAdmin.POST("/", productHandler.Create)
			productAdmin.PUT("/:id", productHandler.Update)
			productAdmin.DELETE("/:id", productHandler.Delete)
		}

		// User & Profile
		api.GET("/admin/users", middleware.RoleMiddleware("admin"), userHandler.GetAll)
		profile := api.Group("/profile")
		{
			profile.GET("/:id", userHandler.GetByID)
			profile.PUT("/:id", userHandler.Update)
			profile.POST("/request-change-email", authHandler.RequestChangeEmail)
			profile.POST("/verify-change-email", authHandler.VerifyChangeEmail)
		}

		// Points & Loyalty
		points := api.Group("/points")
		{
			points.GET("/history", pHandler.GetHistory)
			points.POST("/admin", middleware.RoleMiddleware("admin"), pHandler.CreatePoint)
			points.GET("/admin/all", middleware.RoleMiddleware("admin"), pHandler.GetAllSummaries)
		}

		// Cart System
		cart := api.Group("/cart")
		{
			cart.GET("/", cartHandler.GetMyCart)
			cart.POST("/", cartHandler.AddToCart)
			cart.DELETE("/:id", cartHandler.DeleteItem)
			cart.DELETE("/clear", cartHandler.ClearMyCart)
		}

		// Reward System
		rewards := api.Group("/rewards")
		{
			rewards.GET("/", rewardHandler.GetAll)
			rewards.POST("/claim", rewardHandler.Claim)
			rewards.GET("/history", rewardHandler.GetMyHistory)

			admin := rewards.Group("/admin")
			admin.Use(middleware.RoleMiddleware("admin"))
			{
				admin.GET("/history/all", rewardHandler.GetAllUserHistory)
				admin.POST("/", rewardHandler.Create)
				admin.PUT("/:id", rewardHandler.Update)
				admin.DELETE("/:id", rewardHandler.Delete)
				admin.PATCH("/approve", rewardHandler.Approve)
				admin.PATCH("/reject", rewardHandler.Reject)
			}
		}
	}

	router = r
	log.Println("SUCCESS: Router fully initialized")
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if router == nil {
		http.Error(w, "Service Unavailable: Check Logs", http.StatusServiceUnavailable)
		return
	}
	router.ServeHTTP(w, r)
}
