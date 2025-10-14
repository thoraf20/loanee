package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/thoraf20/loanee/internal/models"
	"gorm.io/gorm"
)

type LoanRepository interface {
	Create(ctx context.Context, loan *models.Loan) error
	FindByUserID(ctx context.Context, userID string) ([]models.Loan, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}

type loanRepository struct {
	db *gorm.DB
}

func NewLoanRepository(db *gorm.DB) LoanRepository {
	return &loanRepository{db: db}
}

func (r *loanRepository) Create(ctx context.Context, loan *models.Loan) error {
	loan.ID = uuid.New()
	loan.CreatedAt = time.Now()
	loan.UpdatedAt = time.Now()

	if err := r.db.WithContext(ctx).Create(loan).Error; err != nil {
		return fmt.Errorf("error adding collateral: %w", err)
	}
	return nil
}

func (r *loanRepository) FindByUserID(ctx context.Context, userID string) ([]models.Loan, error) {
	var loans []models.Loan

	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&loans).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return loans, err
}

func (r *loanRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid loan id: %w", err)
	}

	res := r.db.WithContext(ctx).Model(&models.Loan{}).Where("id = ?", uid).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	})

	if res.Error != nil {
		return fmt.Errorf("error updating loan status: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}