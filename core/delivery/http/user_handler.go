package http

import (
	"errors"
	"io"
	"net/http"
	"ybg-backend-go/core/entity"
	"ybg-backend-go/core/usecase"
	"ybg-backend-go/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserHandler struct {
	uc usecase.UserUsecase
}

func NewUserHandler(uc usecase.UserUsecase) *UserHandler {
	return &UserHandler{uc: uc}
}

func (h *UserHandler) RegisterRoutes(r *gin.Engine) {
	routes := r.Group("/users")
	{
		routes.POST("/", h.Create)
		routes.GET("/", h.GetAll)
		routes.GET("/:id", h.GetByID)
		routes.PUT("/:id", h.Update)
		routes.DELETE("/:id", h.Delete)
		routes.POST("/login", h.Login)
	}
}

func (h *UserHandler) Create(c *gin.Context) {
	var u entity.User
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}
	if err := h.uc.RegisterUser(&u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "User created", "data": u})
}

func (h *UserHandler) GetAll(c *gin.Context) {
	users, err := h.uc.FetchAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(users) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No users found", "data": []entity.User{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}

func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	user, err := h.uc.GetUserProfile(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": user})
}

// func (h *UserHandler) Update(c *gin.Context) {
// 	id, err := uuid.Parse(c.Param("id"))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
// 		return
// 	}

// 	var u entity.User
// 	if err := c.ShouldBindJSON(&u); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// 	u.UserID = id

//		if err := h.uc.UpdateProfile(&u); err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
//			return
//		}
//		c.JSON(http.StatusOK, gin.H{"message": "User updated", "data": u})
//	}
func (h *UserHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	// 1. Binding Data Teks dari Form
	var u entity.User
	u.UserID = id
	u.Name = c.PostForm("name")
	u.Email = c.PostForm("email")

	// 2. Handling File Gambar
	var imageStream io.Reader
	var fileName, contentType string

	file, err := c.FormFile("image")
	if err == nil {
		// Validasi size 5MB
		if file.Size > 5*1024*1024 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Image too large (max 5MB)"})
			return
		}

		openedFile, _ := file.Open()
		defer openedFile.Close()
		imageStream = openedFile
		fileName = file.Filename
		contentType = file.Header.Get("Content-Type")
	}

	// 3. Eksekusi Update
	if err := h.uc.UpdateProfile(&u, imageStream, fileName, contentType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"data": gin.H{
			"user_id":         u.UserID,
			"name":            u.Name,
			"email":           u.Email,
			"profile_picture": u.ProfilePicture,
		},
	})
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	if err := h.uc.RemoveUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	user, err := h.uc.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	token, err := utils.GenerateToken(user.UserID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// TODO: Generate JWT token here and return it
	// c.JSON(http.StatusOK, gin.H{"message": "Login successful", "user": user})
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user": gin.H{ // Kita bentuk ulang objek user-nya di sini
			"user_id": user.UserID,
			"name":    user.Name,
			"email":   user.Email,
			"role":    user.Role,
		},
	})
}
