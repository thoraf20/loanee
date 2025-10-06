package models

import (
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	ID      uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID  uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	AssetType  string    `json:"asset_type" db:"asset_type"` // e.g. ETH, BNB
	Address    string    `json:"address" db:"address"`
	Balance    float64   `json:"balance" db:"balance"`
	IsPrimary  bool      `json:"is_primary" db:"is_primary"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}
