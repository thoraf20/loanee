package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CollateralStatus string

const (
	StatusPending    CollateralStatus = "pending"
	StatusConfirmed  CollateralStatus = "confirmed"
	StatusActive     CollateralStatus = "active"
	StatusReleased   CollateralStatus = "released"
	StatusLiquidated CollateralStatus = "liquidated"
)

type Collateral struct {
	ID             uuid.UUID        `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID         uuid.UUID        `gorm:"type:uuid;not null" json:"user_id"`
	LoanRequestID  uuid.UUID        `gorm:"type:uuid;not null" json:"loan_request_id"`
	AssetSymbol    string           `gorm:"size:10;not null" json:"asset_symbol"`
	AssetAmount    float64          `gorm:"not null" json:"asset_amount"`
	UsdValue       float64          `gorm:"not null" json:"usd_value"`
	FiatCurrency   string           `gorm:"size:5;not null" json:"fiat_currency"`
	Status         CollateralStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	TxHash         *string          `gorm:"size:255" json:"tx_hash"`
	WalletAddress  *string          `gorm:"size:255" json:"wallet_address"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`

	// Optional relationships
	User        User        `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	// LoanRequest LoanRequest `gorm:"foreignKey:LoanRequestID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
}

// BeforeCreate GORM hook — auto-generate UUIDs
func (c *Collateral) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return
}