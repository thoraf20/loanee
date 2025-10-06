package models

import (
	"time"

	"github.com/google/uuid"
)

type UserKYC struct {
	ID      uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID  uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	DocumentType string    `json:"document_type" db:"document_type"` // e.g. passport, id_card
	DocumentURL  string    `json:"document_url" db:"document_url"`   
	Status       string    `json:"status" db:"status"`               // pending, verified, rejected
	SubmittedAt  time.Time `json:"submitted_at" db:"submitted_at"`
	VerifiedAt   *time.Time `json:"verified_at,omitempty" db:"verified_at"`
}