package wallet

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
	Create(ctx context.Context, wallet *models.Wallet) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Wallet, error)
	GetPrimaryByAsset(ctx context.Context, userID uuid.UUID, asset string) (*models.Wallet, error)
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

func (r *repository) Create(ctx context.Context, wallet *models.Wallet) error {
	now := time.Now()
	if wallet.ID == uuid.Nil {
		wallet.ID = uuid.New()
	}
	wallet.CreatedAt = now
	wallet.UpdatedAt = now

	if err := r.db.WithContext(ctx).Create(wallet).Error; err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}
	return nil
}

func (r *repository) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Wallet, error) {
	var wallets []models.Wallet
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_primary DESC, asset_type ASC").
		Find(&wallets).Error; err != nil {
		return nil, fmt.Errorf("failed to list wallets: %w", err)
	}
	return wallets, nil
}

func (r *repository) GetPrimaryByAsset(ctx context.Context, userID uuid.UUID, asset string) (*models.Wallet, error) {
	var wallet models.Wallet
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND asset_type = ?", userID, asset).
		Order("is_primary DESC").
		First(&wallet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query wallet: %w", err)
	}
	return &wallet, nil
}
