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

type CollateralRepository interface {
	Create(ctx context.Context, c *models.Collateral) error
	FindByUserID(ctx context.Context, userID string) ([]models.Collateral, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}

type collateralRepository struct {
	db *gorm.DB
}

func NewCollateralRepository(db *gorm.DB) CollateralRepository {
	return &collateralRepository{db: db}
}

func (r *collateralRepository) Create(ctx context.Context, collateral *models.Collateral) error {
	collateral.ID = uuid.New()
	collateral.CreatedAt = time.Now()
	collateral.UpdatedAt = time.Now()

	if err := r.db.WithContext(ctx).Create(collateral).Error; err != nil {
		return fmt.Errorf("error adding collateral: %w", err)
	}
	return nil
}

func (r *collateralRepository) FindByUserID(ctx context.Context, userID string) ([]models.Collateral, error) {
	var collaterals []models.Collateral
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&collaterals).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return collaterals, err
}

func (r *collateralRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid collateral id: %w", err)
	}

	res := r.db.WithContext(ctx).Model(&models.Collateral{}).Where("id = ?", uid).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	})

	if res.Error != nil {
		return fmt.Errorf("error updating collateral status: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}