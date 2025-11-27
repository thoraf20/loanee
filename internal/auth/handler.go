package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/thoraf20/loanee/internal/utils"
	"github.com/thoraf20/loanee/pkg/response"
	"github.com/thoraf20/loanee/pkg/validator"
)

type Handler struct {
	service   *Service
	validator *validator.Validator
	logger    zerolog.Logger
}

func NewHandler(service *Service, validator *validator.Validator, logger zerolog.Logger) *Handler {
	return &Handler{
		service:   service,
		validator: validator,
		logger:    logger,
	}
}

// Register godoc
// @Summary Register new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration request"
// @Success 201 {object} RegisterResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validator.Validate(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.Register(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error().Err(err).Any("user", &req).Msg("Failed to register user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	utils.Created(c, "registration successful", resp)
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} LoginResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validator.Validate(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error().Err(err).Any("user", &req).Msg("Failed to login user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login user"})
		return
	}

	utils.OK(c, "login successful", resp)
}

// VerifyEmail godoc
// @Summary Verify user email
// @Description Verify user's email address with verification code
// @Tags auth
// @Accept json
// @Produce json
// @Param request body VerifyEmailRequest true "Email verification request"
// @Success 200 {object} VerifyEmailResponse
// @Failure 400 {object} ErrorResponse
// @Router /auth/verify-email [post]
func (h *Handler) VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validator.Validate(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.VerifyEmail(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error().Err(err).Any("user", &req).Msg("Failed to verify user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify user"})
		return
	}

	utils.OK(c, "", resp)
}

// ResendVerificationCode godoc
// @Summary Resend verification code
// @Description Resend email verification code
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ResendCodeRequest true "Resend code request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /auth/resend-code [post]
func (h *Handler) ResendVerificationCode(c *gin.Context) {
	var req ResendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validator.Validate(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.ResendVerificationCode(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error().Err(err).Any("user", &req).Msg("Operation failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Operation failed"})
		return
	}

	utils.OK(c, "", "verification code resend")
}

// ForgotPassword godoc
// @Summary Request password reset
// @Description Request password reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ForgotPasswordRequest true "Forgot password request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /auth/forgot-password [post]
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validator.Validate(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.ForgotPassword(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error().Err(err).Any("user", &req).Msg("Operation failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Operation failed"})
		return
	}

	c.JSON(http.StatusCreated, "password reset code sent")
}

// ResetPassword godoc
// @Summary Reset password
// @Description Reset password with reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Reset password request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /auth/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validator.Validate(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.ResetPassword(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error().Err(err).Any("user", &req).Msg("Operation failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Operation failed"})
		return
	}

	utils.OK(c, "", "password reset successful")
}

// Logout godoc
// @Summary Logout user
// @Description Logout user and invalidate tokens
// @Tags auth
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body LogoutRequest true "Logout request"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	accessToken, exists := c.Get("access_token")
	if !exists {
		utils.BadRequest(c, "Access token not found", nil)
		return
	}

	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	// Logout with both tokens
	err := h.service.Logout(
		c.Request.Context(),
		accessToken.(string),
		req.RefreshToken,
	)
	if err != nil {
		h.logger.Error().Err(err).Any("user", &req).Msg("Failed to logout user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout user"})
		return
	}

	utils.OK(c, "Logged out successfully", nil)
}