package models

import (
	"time"

	"github.com/google/uuid"
)

type UserBalance struct {
	ID      uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID  uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	Currency       string    `json:"currency" db:"currency"` // e.g. USD, NGN
	TotalCollateral float64  `json:"total_collateral" db:"total_collateral"`
	TotalBorrowed   float64  `json:"total_borrowed" db:"total_borrowed"`
	AvailableLimit  float64  `json:"available_limit" db:"available_limit"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}