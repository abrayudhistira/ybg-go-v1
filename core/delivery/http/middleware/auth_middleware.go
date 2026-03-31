package middleware

import (
	"net/http"
	"os"
	"strings"
	"ybg-backend-go/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// func AuthMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		authHeader := c.GetHeader("Authorization")
// 		if authHeader == "" {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
// 			c.Abort()
// 			return
// 		}

// 		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
// 		claims := &utils.Claims{}

// 		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
// 			return []byte("yoursbeyoundglamour"), nil
// 		})

// 		if err != nil || !token.Valid {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
// 			c.Abort()
// 			return
// 		}

//			// Simpan data user ke context agar bisa dipakai di handler
//			c.Set("user_id", claims.UserID)
//			c.Set("role", claims.Role)
//			c.Next()
//		}
//	}
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		claims := &utils.Claims{}

		// Ambil secret key secara dinamis
		jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "yoursbeyoundglamour"
		}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Validasi signing method HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token", "details": err.Error()})
			c.Abort()
			return
		}

		// Karena di Supabase 'role' tersimpan di app_metadata, jika claims.Role kosong
		// Kita bisa beri default 'customer' agar RoleMiddleware tidak error
		userRole := claims.Role
		if userRole == "" {
			userRole = "customer"
		}

		// Simpan data user ke context agar bisa dipakai di handler
		c.Set("user_id", claims.UserID)
		c.Set("role", userRole)
		c.Next()
	}
}

// Middleware khusus untuk cek Role
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Role not found"})
			c.Abort()
			return
		}

		isAllowed := false
		for _, r := range allowedRoles {
			if r == role {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this resource"})
			c.Abort()
			return
		}
		c.Next()
	}
}
