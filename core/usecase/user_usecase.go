package usecase

import (
	"bytes"
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
}

func NewUserUsecase(repo repository.UserRepository, pointRepo repository.PointRepository) UserUsecase {
	return &userUC{repo: repo, pointRepo: pointRepo}
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

// ... Implementasi RegisterUser, Login, dll tetap sama seperti sebelumnya ...
func (u *userUC) RegisterUser(user *entity.User) error {
	if user.UserID == uuid.Nil {
		user.UserID = uuid.New()
	}
	hashed, _ := utils.HashPassword(user.Password)
	user.Password = hashed
	if err := u.repo.Create(user); err != nil {
		return err
	}
	return u.pointRepo.CreatePointTotal(&entity.PointTotal{
		UserID: user.UserID, Total: 0, Tier: "friend", CreatedAt: time.Now(),
	})
}

func (u *userUC) FetchAllUsers() ([]entity.User, error)            { return u.repo.GetAll() }
func (u *userUC) GetUserProfile(id uuid.UUID) (entity.User, error) { return u.repo.GetByID(id) }
func (u *userUC) RemoveUser(id uuid.UUID) error                    { return u.repo.Delete(id) }
func (u *userUC) Login(email, password string) (entity.User, error) {
	user, err := u.repo.GetByEmail(email)
	if err != nil || !utils.CheckPasswordHash(password, user.Password) {
		return entity.User{}, errors.New("invalid credentials")
	}
	return user, nil
}
