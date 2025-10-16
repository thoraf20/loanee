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

type WalletRepository interface {
	Create(ctx context.Context, c *models.Wallet) error
	GetUserWallets(ctx context.Context, userID string) ([]models.Wallet, error)
}

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) WalletRepository {
	if db == nil {
		panic("nil *gorm.DB passed to WalletRepository")
	}
	return &walletRepository{db: db}
}

func (r *walletRepository) Create(ctx context.Context, wallet *models.Wallet) error {
	if wallet == nil {
		return errors.New("cannot create nil wallet")
	}

	wallet.ID = uuid.New()
	now := time.Now()
	wallet.CreatedAt = now
	wallet.UpdatedAt = now

	if err := r.db.WithContext(ctx).Create(wallet).Error; err != nil {
		return fmt.Errorf("error generating wallet: %w", err)
	}
	return nil
}

func (r *walletRepository) GetUserWallets(ctx context.Context, userID string) ([]models.Wallet, error) {
	var wallets []models.Wallet

	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&wallets).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user wallets: %w", err)
	}

	return wallets, err
}