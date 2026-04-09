package usecase

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"ybg-backend-go/core/entity"
	"ybg-backend-go/core/repository"
	"ybg-backend-go/pkg/utils"

	"github.com/google/uuid"
)

type UserUsecase interface {
	RegisterUser(u *entity.User) error
	FetchAllUsers() ([]entity.User, error)
	GetUserProfile(id uuid.UUID) (entity.User, error)
	UpdateProfile(u *entity.User, file io.Reader, fileName, contentType string) error
	RemoveUser(id uuid.UUID) error
	Login(email, password string) (entity.User, error)
}

type userUC struct {
	repo      repository.UserRepository
	pointRepo repository.PointRepository
	authRepo  repository.AuthRepository
}

func NewUserUsecase(
	repo repository.UserRepository,
	pointRepo repository.PointRepository,
	authRepo repository.AuthRepository,
) UserUsecase {
	return &userUC{
		repo:      repo,
		pointRepo: pointRepo,
		authRepo:  authRepo,
	}
}

func (u *userUC) UpdateProfile(user *entity.User, file io.Reader, fileName, contentType string) error {
	// 1. Logika Upload Gambar ke Supabase Storage (jika ada file)
	if file != nil {
		supabaseURL := os.Getenv("SUPABASE_URL")
		supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
		bucketName := "avatars"

		// Path: user_id/filename.ext
		remotePath := fmt.Sprintf("%s/%s", user.UserID.String(), fileName)
		uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, bucketName, remotePath)

		buf := new(bytes.Buffer)
		buf.ReadFrom(file)

		req, _ := http.NewRequest("POST", uploadURL, buf)
		req.Header.Set("Authorization", "Bearer "+supabaseKey)
		req.Header.Set("Content-Type", contentType)
		req.Header.Set("x-upsert", "true") // Overwrite jika sudah ada

		client := &http.Client{}
		resp, err := client.Do(req)
		if err == nil && (resp.StatusCode == http.StatusOK || resp.StatusCode == 201) {
			// Construct Public URL
			user.ProfilePicture = fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, bucketName, remotePath)
		}
		if resp != nil {
			defer resp.Body.Close()
		}
	}

	// 2. Update database
	return u.repo.Update(user)
}

// func (u *userUC) RegisterUser(user *entity.User) error {
// 	existingUser, _ := u.repo.GetByEmail(user.Email)
// 	if existingUser.Email != "" {
// 		return errors.New("email already in use")
// 	}
// 	if user.UserID == uuid.Nil {
// 		user.UserID = uuid.New()
// 	}
// 	hashed, _ := utils.HashPassword(user.Password)
// 	user.Password = hashed
// 	if err := u.repo.Create(user); err != nil {
// 		return err
// 	}
// 	return u.pointRepo.CreatePointTotal(&entity.PointTotal{
// 		UserID: user.UserID, Total: 0, Tier: "friend", CreatedAt: time.Now(),
// 	})
// }

// func (u *userUC) RegisterUser(user *entity.User) error {
// 	// 1. Cek duplikasi email
// 	existingUser, _ := u.repo.GetByEmail(user.Email)
// 	if existingUser.Email != "" {
// 		return errors.New("email already in use")
// 	}

// 	// 2. Set Default ID jika kosong
// 	if user.UserID == uuid.Nil {
// 		user.UserID = uuid.New()
// 	}

// 	// 3. Hash Password (Password sekarang NOT NULL, jadi wajib ada)
// 	if user.Password == "" {
// 		return errors.New("password is required")
// 	}
// 	hashed, _ := utils.HashPassword(user.Password)
// 	user.Password = hashed

// 	// 4. Set Default Role & Time
// 	if user.Role == "" {
// 		user.Role = "customer"
// 	}
// 	user.CreatedAt = time.Now()

// 	// 5. Simpan ke database (Akan error jika Birth, Phone, atau Gender nilainya kosong)
// 	if err := u.repo.Create(user); err != nil {
// 		return fmt.Errorf("failed to create user: %v", err)
// 	}

//		// 6. Inisialisasi Poin
//		return u.pointRepo.CreatePointTotal(&entity.PointTotal{
//			UserID:    user.UserID,
//			Total:     0,
//			Tier:      "friend",
//			CreatedAt: time.Now(),
//		})
//	}
func (u *userUC) RegisterUser(user *entity.User) error {
	// 1. Cek duplikasi email
	existingUser, _ := u.repo.GetByEmail(user.Email)
	if existingUser.Email != "" {
		return errors.New("email sudah terdaftar, silakan gunakan email lain")
	}

	// 2. Persiapan Data User Baru
	if user.UserID == uuid.Nil {
		user.UserID = uuid.New()
	}

	hashed, _ := utils.HashPassword(user.Password)
	user.Password = hashed

	if user.Role == "" {
		user.Role = "customer"
	}

	user.CreatedAt = time.Now()
	user.IsVerified = false // WAJIB: Akun baru dibuat dalam status belum verifikasi

	// 3. Simpan ke Database User
	if err := u.repo.Create(user); err != nil {
		return fmt.Errorf("gagal membuat user: %v", err)
	}

	// 4. Inisialisasi Poin (0 point, Tier Friend)
	err := u.pointRepo.CreatePointTotal(&entity.PointTotal{
		UserID:    user.UserID,
		Total:     0,
		Tier:      "friend",
		CreatedAt: time.Now(),
	})
	if err != nil {
		return fmt.Errorf("gagal inisialisasi poin: %v", err)
	}

	// 5. LOGIKA OTP UNTUK VERIFIKASI AKUN
	// Generate 6 digit OTP
	otp := u.generateRandomOTP(6)

	// Simpan data OTP ke tabel password_resets melalui authRepo
	otpData := &entity.PasswordReset{
		Email:     user.Email,
		OTP:       otp,
		ExpiredAt: time.Now().Add(15 * time.Minute), // Berlaku 15 menit
	}

	if err := u.authRepo.SaveOTP(otpData); err != nil {
		return fmt.Errorf("gagal menyimpan kode verifikasi: %v", err)
	}

	// 6. Kirim Email secara Asynchronous (pakai 'go' agar register tidak terasa lambat)
	// go func() {
	// 	err := utils.SendOTPEmail(user.Email, otp)
	// 	if err != nil {
	// 		fmt.Printf("Gagal kirim email ke %s: %v\n", user.Email, err)
	// 	}
	// }()
	err = utils.SendOTPEmail(user.Email, otp)
	if err != nil {
		fmt.Printf("Gagal kirim email ke %s: %v\n", user.Email, err)
		// Opsi: return fmt.Errorf("user terdaftar tapi email gagal dikirim: %v", err)
	}

	return nil
}

// Helper untuk generate angka (Pastikan method ini ada di dalam user_usecase.go)
func (u *userUC) generateRandomOTP(max int) string {
	var table = []byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	b := make([]byte, max)
	// Kita gunakan math/rand atau crypto/rand yang sudah kamu punya di pkg/utils
	// Jika ingin simpel, panggil fungsi generate yang sudah ada di utils
	n, _ := io.ReadAtLeast(rand.Reader, b, max)
	if n != max {
		return "123456"
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}
func (u *userUC) FetchAllUsers() ([]entity.User, error)            { return u.repo.GetAll() }
func (u *userUC) GetUserProfile(id uuid.UUID) (entity.User, error) { return u.repo.GetByID(id) }
func (u *userUC) RemoveUser(id uuid.UUID) error                    { return u.repo.Delete(id) }

//	func (u *userUC) Login(email, password string) (entity.User, error) {
//		user, err := u.repo.GetByEmail(email)
//		if err != nil || !utils.CheckPasswordHash(password, user.Password) {
//			return entity.User{}, errors.New("invalid credentials")
//		}
//		return user, nil
//	}
func (u *userUC) Login(email, password string) (entity.User, error) {
	user, err := u.repo.GetByEmail(email)
	// 1. Cek keberadaan user
	if err != nil {
		return entity.User{}, errors.New("email atau password salah")
	}

	// 2. Cek status verifikasi (Tambahkan di sini)
	if !user.IsVerified {
		return entity.User{}, errors.New("akun anda belum diverifikasi, silakan cek email untuk kode OTP")
	}

	// 3. Cek password
	if !utils.CheckPasswordHash(password, user.Password) {
		return entity.User{}, errors.New("email atau password salah")
	}

	return user, nil
}
