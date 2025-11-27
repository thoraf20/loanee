package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
  FirstName       string     `gorm:"type:varchar(100);not null" json:"first_name"`
	LastName        string     `gorm:"type:varchar(100);not null" json:"last_name"`
	Email           string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password        string     `gorm:"type:varchar(255);not null" json:"-"`
  PhoneNumber     *string    `gorm:"type:varchar(20)" json:"phone_number,omitempty"`
	Role       			string     `gorm:"type:varchar(100);default:'user'" json:"role"`
	IsVerified      bool       `gorm:"default:false" json:"is_verified"`
	KYCStatus       string     `gorm:"type:varchar(50);default:'pending'" json:"kyc_status"`
	PreferredFiat   string     `gorm:"type:varchar(10);default:'NGN'" json:"preferred_fiat"` // e.g., NGN, USD
	DefaultCurrency string     `gorm:"type:varchar(10);default:'NGN'" json:"default_currency"`
	LastLogin       *time.Time `json:"last_login,omitempty"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

type VerificationCode struct {
	ID        uuid.UUID      `gorm:"primarykey"`
	UserID    uuid.UUID      `gorm:"not null;index"`
	Code      string    		 `gorm:"not null"`
	ExpiresAt time.Time 		 `gorm:"not null;index"`
	Used      bool      		 `gorm:"not null;default:false;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PasswordResetToken struct {
	ID        uuid.UUID      `gorm:"primarykey"`
	UserID    uuid.UUID      `gorm:"not null;index"`
	Token     string    		 `gorm:"not null"`
	ExpiresAt time.Time			 `gorm:"not null;index"`
	Used      bool      		 `gorm:"not null;default:false;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}