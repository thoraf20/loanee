package payment

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
	Create(ctx context.Context, payment *models.Payment) error
	ListByLoan(ctx context.Context, loanID uuid.UUID) ([]models.Payment, error)
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

func (r *repository) Create(ctx context.Context, payment *models.Payment) error {
	now := time.Now()
	if payment.ID == uuid.Nil {
		payment.ID = uuid.New()
	}
	if payment.PaidAt.IsZero() {
		payment.PaidAt = now
	}
	payment.CreatedAt = now
	payment.UpdatedAt = now

	if err := r.db.WithContext(ctx).Create(payment).Error; err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}
	return nil
}

func (r *repository) ListByLoan(ctx context.Context, loanID uuid.UUID) ([]models.Payment, error) {
	var payments []models.Payment
	if err := r.db.WithContext(ctx).
		Where("loan_id = ?", loanID).
		Order("paid_at ASC").
		Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}
	return payments, nil
}
