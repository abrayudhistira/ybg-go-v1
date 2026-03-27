package usecase

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"ybg-backend-go/core/entity"
	"ybg-backend-go/core/repository"
)

type ProductUsecase interface {
	CreateProduct(p *entity.Product, file io.Reader, fileName, contentType string) error
	// FetchProducts() ([]entity.Product, error)
	FetchProducts(search string, limit, offset int) ([]entity.Product, int64, error)
	GetProductDetail(id uint) (entity.Product, error)
	UpdateProduct(p *entity.Product, file io.Reader, fileName, contentType string) error
	DeleteProduct(id uint) error
}

type productUC struct {
	repo repository.ProductRepository
}

func NewProductUsecase(repo repository.ProductRepository) ProductUsecase {
	return &productUC{repo: repo}
}

func (u *productUC) uploadToSupabase(file io.Reader, fileName, contentType string) (string, error) {
	if file == nil {
		return "", nil
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	bucketName := "products"

	// Nama file unik dengan timestamp
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
		return "", fmt.Errorf("failed to upload image to supabase")
	}

	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, bucketName, remotePath), nil
}

func (u *productUC) CreateProduct(p *entity.Product, file io.Reader, fileName, contentType string) error {
	url, err := u.uploadToSupabase(file, fileName, contentType)
	if err == nil && url != "" {
		p.ImageURL = url
	}
	return u.repo.Create(p)
}

func (u *productUC) UpdateProduct(p *entity.Product, file io.Reader, fileName, contentType string) error {
	url, err := u.uploadToSupabase(file, fileName, contentType)
	if err == nil && url != "" {
		p.ImageURL = url
	}
	return u.repo.Update(p)
}

// func (u *productUC) FetchProducts() ([]entity.Product, error)         { return u.repo.GetAll() }
func (u *productUC) FetchProducts(search string, limit, offset int) ([]entity.Product, int64, error) {
	return u.repo.GetAll(search, limit, offset)
}
func (u *productUC) GetProductDetail(id uint) (entity.Product, error) { return u.repo.GetByID(id) }
func (u *productUC) DeleteProduct(id uint) error                      { return u.repo.Delete(id) }
