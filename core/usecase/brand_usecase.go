package usecase

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	"ybg-backend-go/core/entity"
	"ybg-backend-go/core/repository"
)

type BrandUsecase interface {
	CreateBrand(b *entity.Brand, file io.Reader, fileName, contentType string) error
	GetAllBrands() ([]entity.Brand, error)
	DeleteBrand(id uint) error
	UpdateBrand(id uint, b *entity.Brand, file io.Reader, fileName, contentType string) error
}

type brandUC struct {
	repo repository.BrandRepository
}

func NewBrandUsecase(repo repository.BrandRepository) BrandUsecase {
	return &brandUC{repo: repo}
}

// Re-use logika upload dari product (Pastikan Bucket Name sesuai, misal "brands")
func (u *brandUC) uploadToSupabase(file io.Reader, fileName, contentType string) (string, error) {
	if file == nil {
		return "", nil
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	bucketName := "brands" // Sesuaikan nama bucket di Supabase kamu

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

	// Return URL Public
	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, bucketName, remotePath), nil
}

func (u *brandUC) CreateBrand(b *entity.Brand, file io.Reader, fileName, contentType string) error {
	url, err := u.uploadToSupabase(file, fileName, contentType)
	if err == nil && url != "" {
		b.ImageURL = url
	}
	return u.repo.Create(b)
}

func (u *brandUC) GetAllBrands() ([]entity.Brand, error) { return u.repo.GetAll() }

// func (u *brandUC) DeleteBrand(id uint) error             { return u.repo.Delete(id) }
func (u *brandUC) DeleteBrand(id uint) error {
	// 1. Cari data brand-nya dulu buat dapet URL gambarnya
	brand, err := u.repo.GetByID(id)
	if err != nil {
		return err // Brand tidak ditemukan
	}

	// 2. Hapus file di Supabase Storage jika ada
	if brand.ImageURL != "" {
		_ = u.deleteFromSupabase(brand.ImageURL)
		// Kita abaikan error delete storage agar data di DB tetap bisa terhapus
		// walau file storagenya gagal dihapus (misal file sudah hilang duluan)
	}

	// 3. Hapus data di database
	return u.repo.Delete(id)
}

func (u *brandUC) deleteFromSupabase(imageURL string) error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	bucketName := "brands"

	// Potong URL untuk dapetin nama filenya saja
	// URL: .../storage/v1/object/public/brands/123_logo.png
	// Kita butuh: 123_logo.png
	parts := strings.Split(imageURL, "/")
	fileName := parts[len(parts)-1]

	deleteURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, bucketName, fileName)

	req, _ := http.NewRequest("DELETE", deleteURL, nil)
	req.Header.Set("Authorization", "Bearer "+supabaseKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// func (u *brandUC) UpdateBrand(id uint, b *entity.Brand, file io.Reader, fileName, contentType string) error {
// 	// 1. Cek apakah brand-nya ada di DB
// 	// (Opsional: kamu bisa tambah method GetByID di repo kalau mau validasi dulu)

// 	// 2. Jika ada file gambar baru, upload ke Supabase
// 	if file != nil {
// 		url, err := u.uploadToSupabase(file, fileName, contentType)
// 		if err != nil {
// 			return err
// 		}
// 		b.ImageURL = url
// 	}

//		b.BrandID = id          // Pastikan ID-nya sesuai dengan param URL
//		return u.repo.Update(b) // Pastikan di Repo sudah ada method Update
//	}
func (u *brandUC) UpdateBrand(id uint, b *entity.Brand, file io.Reader, fileName, contentType string) error {
	// 1. Ambil data lama dari DB untuk cek URL gambar lama
	oldBrand, err := u.repo.GetByID(id)
	if err != nil {
		return err // Brand tidak ditemukan
	}

	// 2. Jika ada file gambar baru dikirim
	if file != nil {
		// Upload gambar baru
		url, err := u.uploadToSupabase(file, fileName, contentType)
		if err != nil {
			return err
		}

		// HAPUS GAMBAR LAMA (Bug fix: agar storage tidak penuh sampah)
		if oldBrand.ImageURL != "" {
			_ = u.deleteFromSupabase(oldBrand.ImageURL)
		}

		b.ImageURL = url
	}

	b.BrandID = id
	return u.repo.Update(b)
}
