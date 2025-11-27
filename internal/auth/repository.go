package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/thoraf20/loanee/internal/user"
	e "github.com/thoraf20/loanee/pkg/error"
)

// Repository interface defines auth-specific data operations
type Repository interface {
	CreateUser(ctx context.Context, user *user.User) error
	GetUserByEmail(ctx context.Context, email string) (*user.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error)
	UpdateUser(ctx context.Context, user *user.User) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error

	MarkEmailAsVerified(ctx context.Context, userID uuid.UUID) error
	SaveVerificationCode(ctx context.Context, userID uuid.UUID, code string, expiresAt time.Time) error
	GetVerificationCode(ctx context.Context, userID uuid.UUID) (*user.VerificationCode, error)
	InvalidateVerificationCode(ctx context.Context, codeID uuid.UUID, code string) error

	SavePasswordResetToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error
	GetPasswordResetToken(ctx context.Context, token string) (*user.PasswordResetToken, error)
	InvalidatePasswordResetToken(ctx context.Context, token string) error
}

type repository struct {
	db     *gorm.DB
	logger zerolog.Logger
}

func NewRepository(db *gorm.DB, logger zerolog.Logger) Repository {
	return &repository{
		db:     db,
		logger: logger,
	}
}

// CreateUser creates a new user in the database
func (r *repository) CreateUser(ctx context.Context, user *user.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		r.logger.Error().Err(err).Str("email", user.Email).Msg("Failed to create user")
		
		// Check for duplicate entry
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return e.ErrUserEmailExists.WithError(err)
		}
		
		return e.NewDatabaseError("failed to create user", err)
	}

	r.logger.Info().Any("user_id", user.ID).Str("email", user.Email).Msg("User created successfully")
	return nil
}

// GetUserByEmail retrieves a user by email
func (r *repository) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, e.ErrUserNotFound
		}
		r.logger.Error().Err(err).Str("email", email).Msg("Failed to get user by email")
		return nil, e.NewDatabaseError("failed to get user by email", err)
	}

	return &u, nil
}

// GetUserByID retrieves a user by ID
func (r *repository) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	var u user.User
	
	if err := r.db.WithContext(ctx).First(&u, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, e.ErrUserNotFound
		}
		r.logger.Error().Err(err).Any("user_id", id).Msg("Failed to get user by ID")
		return nil, e.NewDatabaseError("failed to get user by ID", err)
	}

	return &u, nil
}

// UpdateUser updates user information
func (r *repository) UpdateUser(ctx context.Context, user *user.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		r.logger.Error().Err(err).Any("user_id", user.ID).Msg("Failed to update user")
		
		// Check for duplicate entry (e.g., email conflict)
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return e.ErrUserEmailExists.WithError(err)
		}
		
		return e.NewDatabaseError("failed to update user", err)
	}

	r.logger.Info().Any("user_id", user.ID).Msg("User updated successfully")
	return nil
}

// UpdatePassword updates a user's password
func (r *repository) UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	result := r.db.WithContext(ctx).
		Model(&user.User{}).
		Where("id = ?", userID).
		Update("password", hashedPassword)

	if result.Error != nil {
		r.logger.Error().Err(result.Error).Any("user_id", userID).Msg("Failed to update password")
		return e.NewDatabaseError("failed to update password", result.Error)
	}

	if result.RowsAffected == 0 {
		return e.ErrUserNotFound
	}

	r.logger.Info().Any("user_id", userID).Msg("Password updated successfully")
	return nil
}

// MarkEmailAsVerified marks a user's email as verified
func (r *repository) MarkEmailAsVerified(ctx context.Context, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&user.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"email_verified": true,
			"verified_at":    time.Now(),
		})

	if result.Error != nil {
		r.logger.Error().Err(result.Error).Any("user_id", userID).Msg("Failed to verify email")
		return e.NewDatabaseError("failed to verify email", result.Error)
	}

	if result.RowsAffected == 0 {
		return e.ErrUserNotFound
	}

	r.logger.Info().Any("user_id", userID).Msg("Email verified successfully")
	return nil
}

// SaveVerificationCode saves a verification code for email verification
func (r *repository) SaveVerificationCode(ctx context.Context, userID uuid.UUID, code string, expiresAt time.Time) error {
	// Hash the code before storing
	hashedCode, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to hash verification code")
		return e.NewInternalError("failed to hash verification code", err)
	}

	verificationCode := &user.VerificationCode{
		UserID:    userID,
		Code:      string(hashedCode),
		ExpiresAt: expiresAt,
		Used:      false,
	}

	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND used = ?", userID, false).
		Delete(&user.VerificationCode{}).Error; err != nil {
		r.logger.Warn().Err(err).Any("user_id", userID).Msg("Failed to delete old verification codes")
	}

	if err := r.db.WithContext(ctx).Create(verificationCode).Error; err != nil {
		r.logger.Error().Err(err).Any("user_id", userID).Msg("Failed to save verification code")
		return e.NewDatabaseError("failed to save verification code", err)
	}

	r.logger.Info().Any("user_id", userID).Msg("Verification code saved successfully")
	return nil
}

// GetVerificationCode retrieves a verification code for a user
func (r *repository) GetVerificationCode(ctx context.Context, userID uuid.UUID) (*user.VerificationCode, error) {
	var code user.VerificationCode
	
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND used = ? AND expires_at > ?", userID, false, time.Now()).
		Order("created_at DESC").
		First(&code).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, e.ErrVerificationCodeNotFound
		}
		r.logger.Error().Err(err).Any("user_id", userID).Msg("Failed to get verification code")
		return nil, e.NewDatabaseError("failed to get verification code", err)
	}

	return &code, nil
}

// ValidateVerificationCode validates and marks a verification code as used
func (r *repository) InvalidateVerificationCode(ctx context.Context, userID uuid.UUID, code string) error {
	verificationCode, err := r.GetVerificationCode(ctx, userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(verificationCode.Code), []byte(code)); err != nil {
		r.logger.Warn().Any("user_id", userID).Msg("Invalid verification code provided")
		return e.ErrVerificationCodeInvalid
	}

	if time.Now().After(verificationCode.ExpiresAt) {
		r.logger.Warn().Any("user_id", userID).Msg("Verification code has expired")
		return e.ErrVerificationCodeExpired
	}

	result := r.db.WithContext(ctx).
		Model(&user.VerificationCode{}).
		Where("id = ?", verificationCode.ID).
		Update("used", true)

	if result.Error != nil {
		r.logger.Error().Err(result.Error).Any("code_id", verificationCode.ID).Msg("Failed to mark verification code as used")
		return e.NewDatabaseError("failed to mark verification code as used", result.Error)
	}

	r.logger.Info().Any("user_id", userID).Msg("Verification code validated successfully")
	return nil
}

// SavePasswordResetToken saves a password reset token
func (r *repository) SavePasswordResetToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	// Hash the token before storing
	hashedToken, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to hash reset token")
		return e.NewInternalError("failed to hash reset token", err)
	}

	resetToken := &user.PasswordResetToken{
		UserID:    userID,
		Token:     string(hashedToken),
		ExpiresAt: expiresAt,
		Used:      false,
	}

	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND used = ?", userID, false).
		Delete(&user.PasswordResetToken{}).Error; err != nil {
		r.logger.Warn().Err(err).Any("user_id", userID).Msg("Failed to delete old reset tokens")
	}

	if err := r.db.WithContext(ctx).Create(resetToken).Error; err != nil {
		r.logger.Error().Err(err).Any("user_id", userID).Msg("Failed to save reset token")
		return e.NewDatabaseError("failed to save reset token", err)
	}

	r.logger.Info().Any("user_id", userID).Msg("Password reset token saved successfully")
	return nil
}

// GetPasswordResetToken retrieves a password reset token
func (r *repository) GetPasswordResetToken(ctx context.Context, token string) (*user.PasswordResetToken, error) {
	var tokens []user.PasswordResetToken
	
	// Get all unused, non-expired tokens
	if err := r.db.WithContext(ctx).
		Where("used = ? AND expires_at > ?", false, time.Now()).
		Find(&tokens).Error; err != nil {
		r.logger.Error().Err(err).Msg("Failed to get reset tokens")
		return nil, e.NewDatabaseError("failed to get reset tokens", err)
	}

	for _, t := range tokens {
		if err := bcrypt.CompareHashAndPassword([]byte(t.Token), []byte(token)); err == nil {
			if time.Now().After(t.ExpiresAt) {
				continue
			}
			return &t, nil
		}
	}

	return nil, e.ErrPasswordResetTokenNotFound
}

// ValidatePasswordResetToken validates a password reset token without marking it as used
func (r *repository) ValidatePasswordResetToken(ctx context.Context, token string) (*user.PasswordResetToken, error) {
	resetToken, err := r.GetPasswordResetToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Additional validation
	if resetToken.Used {
		return nil, e.ErrPasswordResetTokenNotFound.WithDetail("reason", "token already used")
	}

	if time.Now().After(resetToken.ExpiresAt) {
		return nil, e.ErrPasswordResetTokenNotFound.WithDetail("reason", "token expired")
	}

	return resetToken, nil
}

// InvalidatePasswordResetToken marks a password reset token as used
func (r *repository) InvalidatePasswordResetToken(ctx context.Context, token string) error {
	// First get the token
	resetToken, err := r.GetPasswordResetToken(ctx, token)
	if err != nil {
		return err
	}

	// Mark as used
	result := r.db.WithContext(ctx).
		Model(&user.PasswordResetToken{}).
		Where("id = ?", resetToken.ID).
		Update("used", true)

	if result.Error != nil {
		r.logger.Error().Err(result.Error).Any("token_id", resetToken.ID).Msg("Failed to invalidate reset token")
		return e.NewDatabaseError("failed to invalidate reset token", result.Error)
	}

	if result.RowsAffected == 0 {
		return e.ErrPasswordResetTokenNotFound
	}

	r.logger.Info().Any("token_id", resetToken.ID).Msg("Password reset token invalidated")
	return nil
}

// run this in job
func (r *repository) CleanupExpiredVerificationCodes(ctx context.Context) error {
	result := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&user.VerificationCode{})

	if result.Error != nil {
		r.logger.Error().Err(result.Error).Msg("Failed to cleanup expired verification codes")
		return e.NewDatabaseError("failed to cleanup verification codes", result.Error)
	}

	r.logger.Info().Int64("deleted_count", result.RowsAffected).Msg("Cleaned up expired verification codes")
	return nil
}

// put this in job
func (r *repository) CleanupExpiredPasswordResetTokens(ctx context.Context) error {
	result := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&user.PasswordResetToken{})

	if result.Error != nil {
		r.logger.Error().Err(result.Error).Msg("Failed to cleanup expired password reset tokens")
		return e.NewDatabaseError("failed to cleanup password reset tokens", result.Error)
	}

	r.logger.Info().Int64("deleted_count", result.RowsAffected).Msg("Cleaned up expired password reset tokens")
	return nil
}