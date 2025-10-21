package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/thoraf20/loanee/internal/models"
	"gorm.io/gorm"
)

type CollateralRepository interface {
	Create(ctx context.Context, collateral *models.Collateral) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Collateral, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Collateral, error)
	GetByLoanRequestID(ctx context.Context, loanRequestID uuid.UUID) (*models.Collateral, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.CollateralStatus) error
	UpdateTxInfo(ctx context.Context, id uuid.UUID, txHash, walletAddress string) error
}

type collateralRepository struct {
	db *gorm.DB
}

func NewCollateralRepository(db *gorm.DB) CollateralRepository {
	if db == nil {
		panic("nil *gorm.DB passed to CollateralRepository")
	}
	return &collateralRepository{db: db}
}

func (r *collateralRepository) Create(ctx context.Context, collateral *models.Collateral) error {
	collateral.ID = uuid.New()
	collateral.CreatedAt = time.Now()
	collateral.UpdatedAt = time.Now()

	if err := r.db.WithContext(ctx).Create(collateral).Error; err != nil {
		return fmt.Errorf("failed to create collateral: %w", err)
	}
	return nil
}

func (r *collateralRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Collateral, error) {
	var collateral models.Collateral
	if err := r.db.WithContext(ctx).First(&collateral, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get collateral by id: %w", err)
	}
	return &collateral, nil
}

func (r *collateralRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Collateral, error) {
	var collaterals []models.Collateral
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&collaterals).Error; err != nil {
		return nil, fmt.Errorf("failed to get user collaterals: %w", err)
	}
	return collaterals, nil
}

func (r *collateralRepository) GetByLoanRequestID(ctx context.Context, loanRequestID uuid.UUID) (*models.Collateral, error) {
	var collateral models.Collateral
	if err := r.db.WithContext(ctx).Where("loan_request_id = ?", loanRequestID).First(&collateral).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get collateral by loan request: %w", err)
	}
	return &collateral, nil
}

func (r *collateralRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.CollateralStatus) error {
	return r.db.WithContext(ctx).Model(&models.Collateral{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *collateralRepository) UpdateTxInfo(ctx context.Context, id uuid.UUID, txHash, walletAddress string) error {
	updates := map[string]interface{}{
		"tx_hash":        txHash,
		"wallet_address": walletAddress,
		"updated_at":     time.Now(),
	}
	return r.db.WithContext(ctx).Model(&models.Collateral{}).
		Where("id = ?", id).
		Updates(updates).Error
}