package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError represents a custom application error
type AppError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	StatusCode int                    `json:"-"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Err        error                  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithDetail adds a detail field to the error
func (e *AppError) WithDetail(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithError wraps another error
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// NewAppError creates a new application error
func NewAppError(code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Details:    make(map[string]interface{}),
	}
}

// Common error codes
const (
	// Authentication & Authorization
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeForbidden          = "FORBIDDEN"
	CodeTokenInvalid       = "TOKEN_INVALID"
	CodeTokenExpired       = "TOKEN_EXPIRED"
	CodeTokenMissing       = "TOKEN_MISSING"
	CodeInvalidCredentials = "INVALID_CREDENTIALS"
	CodePermissionDenied   = "PERMISSION_DENIED"

	// Validation
	CodeValidationFailed     = "VALIDATION_FAILED"
	CodeInvalidInput         = "INVALID_INPUT"
	CodeMissingField         = "MISSING_FIELD"
	CodeInvalidFormat        = "INVALID_FORMAT"

	// Resource errors
	CodeNotFound           = "NOT_FOUND"
	CodeAlreadyExists      = "ALREADY_EXISTS"
	CodeConflict           = "CONFLICT"
	CodeDependencyConflict = "DEPENDENCY_CONFLICT"

	// Database errors
	CodeDatabaseError      = "DATABASE_ERROR"
	CodeRecordNotFound     = "RECORD_NOT_FOUND"
	CodeDuplicateEntry     = "DUPLICATE_ENTRY"

	// Business logic errors
	CodeInsufficientFunds     = "INSUFFICIENT_FUNDS"
	CodeLTVExceeded          = "LTV_EXCEEDED"
	CodeLoanNotEligible      = "LOAN_NOT_ELIGIBLE"
	CodeCollateralLocked     = "COLLATERAL_LOCKED"
	CodePaymentFailed        = "PAYMENT_FAILED"
	CodeInvalidOperation     = "INVALID_OPERATION"

	// External service errors
	CodeExternalServiceError = "EXTERNAL_SERVICE_ERROR"
	CodePricingServiceError  = "PRICING_SERVICE_ERROR"
	CodePaymentGatewayError  = "PAYMENT_GATEWAY_ERROR"

	// System errors
	CodeInternalServerError = "INTERNAL_SERVER_ERROR"
	CodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
	CodeTimeout             = "TIMEOUT"
)

// Predefined errors

// Authentication & Authorization Errors
var (
	ErrUnauthorized = NewAppError(
		CodeUnauthorized,
		"Authentication required",
		http.StatusUnauthorized,
	)

	ErrForbidden = NewAppError(
		CodeForbidden,
		"You are forbidden from accessing this resource",
		http.StatusForbidden,
	)

	ErrTokenInvalid = NewAppError(
		CodeTokenInvalid,
		"Invalid authentication token",
		http.StatusUnauthorized,
	)

	ErrTokenExpired = NewAppError(
		CodeTokenExpired,
		"Authentication token has expired",
		http.StatusUnauthorized,
	)

	ErrTokenMissing = NewAppError(
		CodeTokenMissing,
		"Authentication token is missing",
		http.StatusUnauthorized,
	)

	ErrInvalidCredentials = NewAppError(
		CodeInvalidCredentials,
		"Invalid email or password",
		http.StatusUnauthorized,
	)

	ErrPermissionDenied = NewAppError(
		CodePermissionDenied,
		"You do not have permission to perform this action",
		http.StatusForbidden,
	)
)

// User Errors
var (
	ErrUserNotFound = NewAppError(
		CodeNotFound,
		"User not found",
		http.StatusNotFound,
	)

	ErrUserExists = NewAppError(
		CodeAlreadyExists,
		"User already exists",
		http.StatusConflict,
	)

	ErrUserEmailExists = NewAppError(
		CodeAlreadyExists,
		"User with this email already exists",
		http.StatusConflict,
	)

	ErrUserPhoneExists = NewAppError(
		CodeAlreadyExists,
		"User with this phone number already exists",
		http.StatusConflict,
	)

	ErrEmailNotVerified = NewAppError(
		CodeInvalidOperation,
		"Email address is not verified",
		http.StatusForbidden,
	)
)

// Verification Errors
var (
	ErrVerificationCodeInvalid = NewAppError(
		CodeInvalidInput,
		"Invalid verification code",
		http.StatusBadRequest,
	)

	ErrVerificationCodeExpired = NewAppError(
		CodeTokenExpired,
		"Verification code has expired",
		http.StatusBadRequest,
	)

	ErrVerificationCodeNotFound = NewAppError(
		CodeNotFound,
		"Verification code not found",
		http.StatusNotFound,
	)

	ErrPasswordResetTokenNotFound = NewAppError(
		CodeNotFound,
		"Password reset token not found or expired",
		http.StatusNotFound,
	)
)

// Collateral Errors
var (
	ErrCollateralNotFound = NewAppError(
		CodeNotFound,
		"Collateral not found",
		http.StatusNotFound,
	)

	ErrCollateralLocked = NewAppError(
		CodeConflict,
		"Collateral is locked and cannot be modified",
		http.StatusConflict,
	)

	ErrCollateralInsufficientValue = NewAppError(
		CodeInvalidOperation,
		"Collateral value is insufficient for the requested loan",
		http.StatusBadRequest,
	)

	ErrCollateralAlreadyExists = NewAppError(
		CodeAlreadyExists,
		"Collateral with this asset already exists",
		http.StatusConflict,
	)
)

// Loan Errors
var (
	ErrLoanNotFound = NewAppError(
		CodeNotFound,
		"Loan not found",
		http.StatusNotFound,
	)

	ErrLoanNotEligible = NewAppError(
		CodeLoanNotEligible,
		"User is not eligible for this loan",
		http.StatusBadRequest,
	)

	ErrLTVExceeded = NewAppError(
		CodeLTVExceeded,
		"Loan amount exceeds maximum LTV ratio",
		http.StatusBadRequest,
	)

	ErrLoanAlreadyActive = NewAppError(
		CodeConflict,
		"An active loan already exists for this collateral",
		http.StatusConflict,
	)

	ErrLoanNotActive = NewAppError(
		CodeInvalidOperation,
		"Loan is not active",
		http.StatusBadRequest,
	)
)

// Payment Errors
var (
	ErrPaymentNotFound = NewAppError(
		CodeNotFound,
		"Payment not found",
		http.StatusNotFound,
	)

	ErrPaymentFailed = NewAppError(
		CodePaymentFailed,
		"Payment processing failed",
		http.StatusBadRequest,
	)

	ErrInsufficientPaymentAmount = NewAppError(
		CodeInvalidInput,
		"Payment amount is insufficient",
		http.StatusBadRequest,
	)

	ErrPaymentAlreadyProcessed = NewAppError(
		CodeConflict,
		"Payment has already been processed",
		http.StatusConflict,
	)
)

// Database Errors
var (
	ErrDatabaseOperation = NewAppError(
		CodeDatabaseError,
		"Database operation failed",
		http.StatusInternalServerError,
	)

	ErrRecordNotFound = NewAppError(
		CodeRecordNotFound,
		"Record not found",
		http.StatusNotFound,
	)

	ErrDuplicateEntry = NewAppError(
		CodeDuplicateEntry,
		"Duplicate entry",
		http.StatusConflict,
	)

	ErrDependencyConflict = NewAppError(
		CodeDependencyConflict,
		"Resource cannot be deleted until dependent resources are removed",
		http.StatusConflict,
	)
)

// External Service Errors
var (
	ErrPricingServiceUnavailable = NewAppError(
		CodePricingServiceError,
		"Pricing service is currently unavailable",
		http.StatusServiceUnavailable,
	)

	ErrPriceFetchFailed = NewAppError(
		CodePricingServiceError,
		"Failed to fetch asset price",
		http.StatusBadGateway,
	)

	ErrExternalServiceError = NewAppError(
		CodeExternalServiceError,
		"External service error",
		http.StatusBadGateway,
	)
)

// Validation Errors
var (
	ErrValidationFailed = NewAppError(
		CodeValidationFailed,
		"Validation failed",
		http.StatusBadRequest,
	)

	ErrInvalidInput = NewAppError(
		CodeInvalidInput,
		"Invalid input provided",
		http.StatusBadRequest,
	)

	ErrMissingRequiredField = NewAppError(
		CodeMissingField,
		"Required field is missing",
		http.StatusBadRequest,
	)
)

// System Errors
var (
	ErrInternalServer = NewAppError(
		CodeInternalServerError,
		"Internal server error",
		http.StatusInternalServerError,
	)

	ErrServiceUnavailable = NewAppError(
		CodeServiceUnavailable,
		"Service temporarily unavailable",
		http.StatusServiceUnavailable,
	)

	ErrTimeout = NewAppError(
		CodeTimeout,
		"Request timeout",
		http.StatusGatewayTimeout,
	)
)

// Helper functions

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError extracts AppError from an error
func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}

// NewDatabaseError creates a database error with wrapped error
func NewDatabaseError(message string, err error) *AppError {
	return &AppError{
		Code:       CodeDatabaseError,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// NewValidationError creates a validation error with field details
func NewValidationError(message string, fieldErrors map[string]string) *AppError {
	appErr := NewAppError(CodeValidationFailed, message, http.StatusBadRequest)
	for field, errMsg := range fieldErrors {
		appErr.WithDetail(field, errMsg)
	}
	return appErr
}

// NewNotFoundError creates a not found error for a resource
func NewNotFoundError(resource string) *AppError {
	return NewAppError(
		CodeNotFound,
		fmt.Sprintf("%s not found", resource),
		http.StatusNotFound,
	)
}

// NewConflictError creates a conflict error
func NewConflictError(resource, reason string) *AppError {
	return NewAppError(
		CodeConflict,
		fmt.Sprintf("%s conflict: %s", resource, reason),
		http.StatusConflict,
	)
}

// NewUnauthorizedError creates an unauthorized error with custom message
func NewUnauthorizedError(message string) *AppError {
	return NewAppError(CodeUnauthorized, message, http.StatusUnauthorized)
}

// NewForbiddenError creates a forbidden error with custom message
func NewForbiddenError(message string) *AppError {
	return NewAppError(CodeForbidden, message, http.StatusForbidden)
}

// NewBadRequestError creates a bad request error
func NewBadRequestError(message string) *AppError {
	return NewAppError(CodeInvalidInput, message, http.StatusBadRequest)
}

// NewInternalError creates an internal server error with wrapped error
func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Code:       CodeInternalServerError,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// WrapError wraps a standard error into an AppError
func WrapError(err error, code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}