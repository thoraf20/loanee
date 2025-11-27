package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CollateralStatus string

const (
	StatusPreview          CollateralStatus = "preview"
	StatusPending          CollateralStatus = "pending"
	StatusConfirmed        CollateralStatus = "confirmed"
	StatusActive           CollateralStatus = "active"
	StatusReleaseRequested CollateralStatus = "release_requested"
	StatusReleased         CollateralStatus = "released"
	StatusLiquidated       CollateralStatus = "liquidated"
)

type Collateral struct {
	ID                 uuid.UUID        `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID             uuid.UUID        `gorm:"type:uuid;not null" json:"user_id"`
	LoanRequestID      *uuid.UUID       `gorm:"type:uuid" json:"loan_request_id,omitempty"`
	AssetSymbol        string           `gorm:"size:10;not null" json:"asset_symbol"`
	AssetAmount        float64          `gorm:"not null" json:"asset_amount"`
	AssetValue         float64          `gorm:"not null" json:"asset_value"`
	RequiredValue      float64          `gorm:"not null" json:"required_value"`
	FiatCurrency       string           `gorm:"size:5;not null" json:"fiat_currency"`
	FiatAmount         float64          `gorm:"not null" json:"fiat_amount"`
	LTV                float64          `gorm:"not null;default:0.65" json:"ltv"`
	Status             CollateralStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	TxHash             *string          `gorm:"size:255" json:"tx_hash,omitempty"`
	WalletAddress      *string          `gorm:"size:255" json:"wallet_address,omitempty"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
	VerifiedAt         *time.Time       `json:"verified_at,omitempty"`
	ReleaseRequestedAt *time.Time       `json:"release_requested_at,omitempty"`
	ReleaseResolvedAt  *time.Time       `json:"release_resolved_at,omitempty"`
	ReleaseNote        *string          `json:"release_note,omitempty"`
}

// BeforeCreate GORM hook — auto-generate UUIDs
func (c *Collateral) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return
}
