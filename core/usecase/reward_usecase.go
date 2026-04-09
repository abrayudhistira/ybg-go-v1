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

	"github.com/google/uuid"
)

type RewardUsecase interface {
	GetAllRewards() ([]entity.Reward, error)
	ClaimReward(userID uuid.UUID, rewardID uuid.UUID) error
	GetMyHistory(userID uuid.UUID) ([]entity.RewardHistory, error)
	ApproveClaim(historyID uuid.UUID, note string) error
	RejectClaim(historyID uuid.UUID, reason string) error
	CreateReward(reward *entity.Reward, file io.Reader, fileName, contentType string) error
	UpdateReward(reward *entity.Reward, file io.Reader, fileName, contentType string) error
	DeleteReward(id uuid.UUID) error
	GetAllUserHistory(page, size int) ([]entity.RewardHistory, int64, error)
}

type rewardUC struct {
	repo      repository.RewardRepository
	pointRepo repository.PointRepository
}

func NewRewardUsecase(r repository.RewardRepository, p repository.PointRepository) RewardUsecase {
	return &rewardUC{repo: r, pointRepo: p}
}

func (u *rewardUC) GetAllRewards() ([]entity.Reward, error) {
	return u.repo.GetAll()
}

// func (u *rewardUC) ClaimReward(userID uuid.UUID, rewardID uuid.UUID) error {
// 	// 1. Ambil detail reward & cek stok
// 	reward, err := u.repo.GetByID(rewardID)
// 	if err != nil {
// 		return errors.New("reward tidak ditemukan")
// 	}
// 	if reward.Quantity <= 0 {
// 		return errors.New("stok reward sudah habis")
// 	}

// 	// 2. Cek saldo poin user (Sekarang sudah aman karena GetPointTotal sudah ada)
// 	pointTotal, err := u.pointRepo.GetPointTotal(userID)
// 	if err != nil {
// 		return errors.New("data poin user tidak ditemukan")
// 	}
// 	if pointTotal.Total < reward.PointCost {
// 		return errors.New("poin tidak cukup")
// 	}

// 	// 3. Eksekusi Pengurangan (Stok & Poin Total)
// 	// Di sini kita kirim MINUS pointCost karena UpdateTotal di repo pakai gorm.Expr(total + addedPoint)
// 	if err := u.pointRepo.UpdateTotal(userID, -reward.PointCost); err != nil {
// 		return errors.New("gagal memproses pengurangan poin")
// 	}

// 	if err := u.repo.UpdateQuantity(rewardID, reward.Quantity-1); err != nil {
// 		return errors.New("gagal mengupdate stok")
// 	}

// 	// 4. Catat di Point History
// 	pointHistory := &entity.PointHistory{
// 		UserID: userID,
// 		Point:  -reward.PointCost,
// 		Status: "aktif",
// 	}
// 	_ = u.pointRepo.CreateHistory(pointHistory)

// 	// 5. Catat di Reward History
// 	claim := &entity.RewardHistory{
// 		UserID:     userID,
// 		RewardID:   rewardID,
// 		PointSpent: reward.PointCost,
// 		Status:     "pengajuan",
// 	}
// 	return u.repo.CreateHistory(claim)
// }

func (u *rewardUC) ClaimReward(userID uuid.UUID, rewardID uuid.UUID) error {
	reward, err := u.repo.GetByID(rewardID)
	if err != nil {
		return errors.New("reward tidak ditemukan")
	}

	if reward.Quantity <= 0 {
		return errors.New("stok habis")
	}

	// Cek saldo awal (biar gak asal klaim)
	pointTotal, _ := u.pointRepo.GetPointTotal(userID)
	if pointTotal.Total < reward.PointCost {
		return errors.New("poin tidak cukup untuk mengajukan klaim")
	}

	claim := &entity.RewardHistory{
		UserID:     userID,
		RewardID:   rewardID,
		PointSpent: reward.PointCost,
		Status:     "pengajuan", // Masuk ke antrian
	}
	return u.repo.CreateHistory(claim)
}

func (u *rewardUC) GetMyHistory(userID uuid.UUID) ([]entity.RewardHistory, error) {
	return u.repo.GetHistoryByUserID(userID)
}
func (u *rewardUC) ApproveClaim(historyID uuid.UUID, note string) error {
	// 1. Ambil data history klaim
	history, err := u.repo.GetHistoryByID(historyID)
	if err != nil {
		return err
	}
	if history.Status != "pengajuan" {
		return errors.New("hanya status pengajuan yang bisa di-acc")
	}

	// 2. Ambil detail reward & user
	reward, _ := u.repo.GetByID(history.RewardID)
	pointTotal, _ := u.pointRepo.GetPointTotal(history.UserID)

	// 3. Validasi Akhir sebelum potong
	if reward.Quantity <= 0 {
		return errors.New("stok mendadak habis")
	}
	if pointTotal.Total < history.PointSpent {
		return errors.New("poin user sudah tidak cukup")
	}

	// 4. EKSEKUSI (Potong Poin & Stok)
	_ = u.pointRepo.UpdateTotal(history.UserID, -history.PointSpent)
	_ = u.repo.UpdateQuantity(history.RewardID, reward.Quantity-1)

	// 5. Update Status ke "acc" dengan Note
	if err := u.repo.UpdateHistoryStatus(historyID, "acc", note); err != nil {
		return err
	}

	// 6. Catat di Point History sebagai poin keluar (status: 'used' atau 'aktif')
	pointEntry := &entity.PointHistory{
		UserID: history.UserID,
		Point:  -history.PointSpent,
		Status: "aktif",
	}
	return u.pointRepo.CreateHistory(pointEntry)
}
func (u *rewardUC) RejectClaim(historyID uuid.UUID, reason string) error {
	// 1. Ambil data history
	history, err := u.repo.GetHistoryByID(historyID)
	if err != nil {
		return err
	}

	// 2. Pastikan hanya yang statusnya 'pengajuan' yang bisa ditolak
	if history.Status != "pengajuan" {
		return errors.New("hanya pengajuan aktif yang bisa ditolak")
	}

	// 3. Update status jadi ditolak
	// Kita bisa manfaatkan kolom admin_note untuk isi alasan penolakan
	return u.repo.UpdateHistoryStatus(historyID, "ditolak", reason)
}
func (u *rewardUC) uploadToSupabase(file io.Reader, fileName, contentType string) (string, error) {
	if file == nil {
		return "", nil
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	bucketName := "rewards" // Pastikan bucket ini ada di Supabase

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
		return "", fmt.Errorf("failed to upload reward image")
	}

	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, bucketName, remotePath), nil
}

func (u *rewardUC) CreateReward(reward *entity.Reward, file io.Reader, fileName, contentType string) error {
	// Upload jika ada file
	url, err := u.uploadToSupabase(file, fileName, contentType)
	if err == nil && url != "" {
		reward.ImageURL = url
	}

	if reward.RewardID == uuid.Nil {
		reward.RewardID = uuid.New()
	}
	return u.repo.Create(reward)
}
func (u *rewardUC) UpdateReward(reward *entity.Reward, file io.Reader, fileName, contentType string) error {
	// Jika ada file baru, upload dan ganti ImageURL
	url, err := u.uploadToSupabase(file, fileName, contentType)
	if err == nil && url != "" {
		reward.ImageURL = url
	}
	return u.repo.Update(reward)
}

func (u *rewardUC) DeleteReward(id uuid.UUID) error {
	return u.repo.Delete(id)
}
func (u *rewardUC) GetAllUserHistory(page, size int) ([]entity.RewardHistory, int64, error) {
	// Default nilai jika kosong
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}

	offset := (page - 1) * size
	return u.repo.GetAllHistories(size, offset)
}
