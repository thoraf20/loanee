package models

import (
	"time"

	"github.com/google/uuid"
)

type Loan struct {
	ID              uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID          uuid.UUID
	CollateralID    uuid.UUID
	AmountRequested float64
	AmountApproved  float64
	InterestRate    float64
	DurationMonths  int
	Status          string    `gorm:"default:'pending'"` // pending, approved, active, repaid, defaulted
	CreatedAt       time.Time
	UpdatedAt       time.Time
}