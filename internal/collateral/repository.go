package collateral

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
	Create(ctx context.Context, collateral *models.Collateral) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Collateral, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Collateral, error)
	ListAll(ctx context.Context) ([]models.Collateral, error)
	Update(ctx context.Context, collateral *models.Collateral) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.CollateralStatus) error
	UpdateTxInfo(ctx context.Context, id uuid.UUID, txHash, walletAddress string, status models.CollateralStatus) error
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

func (r *repository) Create(ctx context.Context, collateral *models.Collateral) error {
	collateral.ID = uuid.New()
	now := time.Now()
	collateral.CreatedAt = now
	collateral.UpdatedAt = now

	if err := r.db.WithContext(ctx).Create(collateral).Error; err != nil {
		r.logger.Error().Err(err).Msg("failed to create collateral record")
		return fmt.Errorf("failed to create collateral: %w", err)
	}
	return nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*models.Collateral, error) {
	var collateral models.Collateral
	if err := r.db.WithContext(ctx).First(&collateral, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.logger.Error().Err(err).Any("id", id).Msg("failed to fetch collateral by id")
		return nil, fmt.Errorf("failed to fetch collateral: %w", err)
	}
	return &collateral, nil
}

func (r *repository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Collateral, error) {
	var collaterals []models.Collateral
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&collaterals).Error; err != nil {
		r.logger.Error().Err(err).Any("user_id", userID).Msg("failed to fetch user collaterals")
		return nil, fmt.Errorf("failed to query collaterals: %w", err)
	}
	return collaterals, nil
}

func (r *repository) ListAll(ctx context.Context) ([]models.Collateral, error) {
	var collaterals []models.Collateral
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&collaterals).Error; err != nil {
		return nil, fmt.Errorf("failed to list collaterals: %w", err)
	}
	return collaterals, nil
}

func (r *repository) Update(ctx context.Context, collateral *models.Collateral) error {
	collateral.UpdatedAt = time.Now()
	if err := r.db.WithContext(ctx).Save(collateral).Error; err != nil {
		return fmt.Errorf("failed to update collateral: %w", err)
	}
	return nil
}

func (r *repository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.CollateralStatus) error {
	res := r.db.WithContext(ctx).
		Model(&models.Collateral{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return fmt.Errorf("failed to update status: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *repository) UpdateTxInfo(ctx context.Context, id uuid.UUID, txHash, walletAddress string, status models.CollateralStatus) error {
	res := r.db.WithContext(ctx).
		Model(&models.Collateral{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"tx_hash":        txHash,
			"wallet_address": walletAddress,
			"status":         status,
			"updated_at":     time.Now(),
			"verified_at":    time.Now(),
		})
	if res.Error != nil {
		return fmt.Errorf("failed to update tx info: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
