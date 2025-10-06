package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID      uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	FirstName       string     `json:"first_name" db:"first_name"`
	LastName        string     `json:"last_name" db:"last_name"`
	Email           string     `json:"email" db:"email"`
	Password        string     `json:"-" db:"password"`
	PhoneNumber     string     `json:"phone_number" db:"phone_number"`
	IsVerified      bool       `json:"is_verified" db:"is_verified"`
	KYCStatus       string     `json:"kyc_status" db:"kyc_status"` // pending|verified|rejected
	PreferredFiat   string     `json:"preferred_fiat" db:"preferred_fiat"` // e.g. NGN, USD
	DefaultCurrency string     `json:"default_currency" db:"default_currency"`
	LastLogin       *time.Time `json:"last_login,omitempty" db:"last_login"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}