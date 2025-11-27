package models

import (
	"time"

	"github.com/google/uuid"
)

type Loan struct {
	ID                   uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID               uuid.UUID `gorm:"type:uuid;not null"`
	CollateralID         uuid.UUID `gorm:"type:uuid;not null"`
	AmountRequested      float64   `gorm:"not null"`
	AmountApproved       float64   `gorm:"not null"`
	PrincipalOutstanding float64   `gorm:"not null"`
	InterestRate         float64   `gorm:"not null"`
	DurationMonths       int       `gorm:"not null"`
	DisbursedAt          *time.Time
	NextDueDate          *time.Time
	TotalRepaid          float64 `gorm:"not null;default:0"`
	PenaltyAccrued       float64 `gorm:"not null;default:0"`
	LastPaymentAt        *time.Time
	Status               string `gorm:"type:varchar(20);default:'pending'"` // pending, approved, disbursed, active, repaid, defaulted
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
