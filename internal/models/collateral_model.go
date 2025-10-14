package models

import (
	"time"

	"github.com/google/uuid"
)

type Collateral struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID       uuid.UUID `gorm:"type:uuid;not null"`
	Asset        string    `gorm:"type:varchar(10);not null"` // BTC, ETH, USDT
	Amount       float64   `gorm:"not null"`                  // amount locked
	LockedValue  float64   `gorm:"not null"`                  // fiat equivalent
	Status       string    `gorm:"default:'locked'"`          // locked, released, liquidated
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
