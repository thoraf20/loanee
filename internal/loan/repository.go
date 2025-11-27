package loan

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/thoraf20/loanee/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, loan *models.Loan) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Loan, error)
	ListAll(ctx context.Context) ([]models.Loan, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Loan, error)
	Update(ctx context.Context, loan *models.Loan) error
}

type repository struct {
	db     *gorm.DB
	logger zerolog.Logger
}

func NewRepository(db *gorm.DB, logger zerolog.Logger) Repository {
	return &repository{
		db:     db,
		logger: logger,
	}
}

func (r *repository) Create(ctx context.Context, loan *models.Loan) error {
	now := time.Now()
	if loan.ID == uuid.Nil {
		loan.ID = uuid.New()
	}
	loan.CreatedAt = now
	loan.UpdatedAt = now
	if err := r.db.WithContext(ctx).Create(loan).Error; err != nil {
		return fmt.Errorf("failed to create loan: %w", err)
	}
	return nil
}

func (r *repository) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Loan, error) {
	var loans []models.Loan
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&loans).Error; err != nil {
		return nil, fmt.Errorf("failed to list loans: %w", err)
	}
	return loans, nil
}

func (r *repository) ListAll(ctx context.Context) ([]models.Loan, error) {
	var loans []models.Loan
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&loans).Error; err != nil {
		return nil, fmt.Errorf("failed to list all loans: %w", err)
	}
	return loans, nil
}

func (r *repository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	res := r.db.WithContext(ctx).
		Model(&models.Loan{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return fmt.Errorf("failed to update loan status: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*models.Loan, error) {
	var loan models.Loan
	if err := r.db.WithContext(ctx).First(&loan, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get loan: %w", err)
	}
	return &loan, nil
}

func (r *repository) Update(ctx context.Context, loan *models.Loan) error {
	loan.UpdatedAt = time.Now()
	if err := r.db.WithContext(ctx).Save(loan).Error; err != nil {
		return fmt.Errorf("failed to update loan: %w", err)
	}
	return nil
}
