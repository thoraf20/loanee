package models

import (
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	ID              uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID  				uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	AssetType  			string    `gorm:"asset_type" db:"asset_type"` // e.g. ETH, BNB
	Address    			string    `gorm:"address" db:"address"`
	PrivateKey  		string    `gorm:"type:text" json:"-"`
	Balance    			float64   `gorm:"balance" db:"balance"`
	IsPrimary  			bool      `gorm:"is_primary" db:"is_primary"`
	CreatedAt   		time.Time
	UpdatedAt   		time.Time
}