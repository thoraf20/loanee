package models

import (
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentCompleted PaymentStatus = "completed"
	PaymentFailed    PaymentStatus = "failed"
)

type Payment struct {
	ID              uuid.UUID     `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	LoanID          uuid.UUID     `gorm:"type:uuid;not null" json:"loan_id"`
	UserID          uuid.UUID     `gorm:"type:uuid;not null" json:"user_id"`
	Amount          float64       `gorm:"not null" json:"amount"`
	Currency        string        `gorm:"size:5;not null" json:"currency"`
	PrincipalAmount float64       `gorm:"not null" json:"principal_amount"`
	InterestAmount  float64       `gorm:"not null" json:"interest_amount"`
	PenaltyAmount   float64       `gorm:"not null" json:"penalty_amount"`
	Method          string        `gorm:"size:50" json:"method"`
	Reference       string        `gorm:"size:100" json:"reference"`
	Status          PaymentStatus `gorm:"type:varchar(20);default:'completed'" json:"status"`
	PaidAt          time.Time     `json:"paid_at"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}
