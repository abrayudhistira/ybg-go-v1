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
)

type NewsUsecase interface {
	CreateNews(n *entity.News, file io.Reader, fileName, contentType string) error
	GetAllNews() ([]entity.News, error)
	GetNewsByID(id uint) (entity.News, error)
	UpdateNews(n *entity.News, file io.Reader, fileName, contentType string) error
	DeleteNews(id uint) error
}

type newsUC struct {
	repo repository.NewsRepository
}

func NewNewsUsecase(repo repository.NewsRepository) NewsUsecase {
	return &newsUC{repo: repo}
}

func (u *newsUC) CreateNews(n *entity.News, file io.Reader, fileName, contentType string) error {
	if file != nil {
		imgURL, err := u.uploadToSupabase(file, fileName, contentType)
		if err != nil {
			return err
		}
		n.ImageURL = imgURL
	}
	return u.repo.Create(n)
}

// func (u *newsUC) UpdateNews(n *entity.News, file io.Reader, fileName, contentType string) error {
// 	if file != nil {
// 		imgURL, err := u.uploadToSupabase(file, fileName, contentType)
// 		if err != nil {
// 			return err
// 		}
// 		n.ImageURL = imgURL
// 	}
// 	return u.repo.Update(n)
// }

func (u *newsUC) UpdateNews(n *entity.News, file io.Reader, fileName, contentType string) error {
	// 1. Ambil data lama dari database untuk mendapatkan ImageURL yang sudah ada
	existingNews, err := u.repo.GetByID(n.NewsID)
	if err != nil {
		return err
	}

	// 2. Jika ada file baru, upload dan ganti URL-nya
	if file != nil {
		imgURL, err := u.uploadToSupabase(file, fileName, contentType)
		if err != nil {
			return err
		}
		n.ImageURL = imgURL
	} else {
		// 3. JIKA TIDAK ADA FILE BARU, gunakan ImageURL yang lama
		n.ImageURL = existingNews.ImageURL
	}

	return u.repo.Update(n)
}

// Helper Upload
func (u *newsUC) uploadToSupabase(file io.Reader, fileName, contentType string) (string, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	bucketName := "news" // Pastikan bucket ini ada di Supabase

	remotePath := fmt.Sprintf("%d_%s", time.Now().Unix(), fileName)
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, bucketName, remotePath)

	buf := new(bytes.Buffer)
	buf.ReadFrom(file)

	req, _ := http.NewRequest("POST", uploadURL, buf)
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != 201 {
		return "", errors.New("gagal upload gambar ke supabase")
	}

	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, bucketName, remotePath), nil
}

func (u *newsUC) GetAllNews() ([]entity.News, error)       { return u.repo.GetAll() }
func (u *newsUC) GetNewsByID(id uint) (entity.News, error) { return u.repo.GetByID(id) }
func (u *newsUC) DeleteNews(id uint) error                 { return u.repo.Delete(id) }
