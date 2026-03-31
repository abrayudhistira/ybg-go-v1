package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Gunakan environment variable, jika kosong baru pakai fallback string
func getJwtKey() []byte {
	secret := os.Getenv("SUPABASE_JWT_SECRET")
	if secret == "" {
		return []byte("yoursbeyoundglamour")
	}
	return []byte(secret)
}

type Claims struct {
	// Kita gunakan string untuk UserID agar kompatibel dengan 'sub' dari Supabase
	// Tag 'sub' wajib ada agar bisa membaca token dari Google OAuth Supabase
	UserID string `json:"sub"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(userID uuid.UUID, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID.String(), // Convert UUID ke string
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Subject:   userID.String(), // Set juga di field standar 'sub'
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJwtKey())
}
