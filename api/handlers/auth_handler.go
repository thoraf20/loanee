package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/thoraf20/loanee/internal/dtos"
	"github.com/thoraf20/loanee/internal/services"
	"github.com/thoraf20/loanee/internal/utils"
	"github.com/thoraf20/loanee/pkg/binding"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	req, verr := binding.StrictBindJSON[dtos.RegisterUserDTO](r)

	if verr != nil {
		utils.Error(w, http.StatusBadRequest, verr.Message, verr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	user, err := h.authService.RegisterUser(ctx, *req)
	if err != nil {
		log.Error().Err(err).Msg("failed to register user")
		utils.Error(w, http.StatusConflict, "Registration failed", err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, "User registered successfully", user)
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	req, verr := binding.StrictBindJSON[dtos.VerifyEmailDTO](r)

	if verr != nil {
		utils.Error(w, http.StatusBadRequest, verr.Message, verr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	verify, err := h.authService.VerifyEmail(ctx, *req)
	if err != nil {
		log.Error().Err(err).Msg("failed to verify email")
		utils.Error(w, http.StatusBadRequest, "Email verification failed", err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, "Email verification successfully", verify)
}

// LoginUser handles user login and JWT generation
func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	req, verr := binding.StrictBindJSON[dtos.LoginDTO](r)

	if verr != nil {
		utils.Error(w, http.StatusBadRequest, verr.Message, verr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	token, err := h.authService.LoginUser(ctx, *req)
	if err != nil {
		log.Warn().Err(err).Msg("failed login attempt")
		utils.Error(w, http.StatusUnauthorized, "Invalid credentials", err.Error())
		return
	}

	response := map[string]string{ "token": token }
	utils.JSON(w, http.StatusOK, "Login successful", response)
}

// @Summary Request password reset
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body auth.PasswordResetRequest true "Password reset info"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /auth/password/reset/request [post]
// handlers/auth_handler.go
func (h *AuthHandler) PasswordResetRequest(w http.ResponseWriter, r *http.Request) {
	req, verr := binding.StrictBindJSON[dtos.PasswordRequestDTO](r)

	if verr != nil {
		utils.Error(w, http.StatusBadRequest, verr.Message, verr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	data, err := h.authService.PasswordResetRequest(ctx, *req)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "Password reset failed", err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, "Password reset request processed", data)
}


// @Summary Reset reset
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body auth.PasswordResetRequest true "Reset password info"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /auth/password/reset [post]
func (h *AuthHandler) PasswordReset(w http.ResponseWriter, r *http.Request) {
	req, verr := binding.StrictBindJSON[dtos.PasswordResetDTO](r)

	if verr != nil {
		utils.Error(w, http.StatusBadRequest, verr.Message, verr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	data, err := h.authService.PasswordReset(ctx, *req)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "Password reset failed", err.Error())
		return
	}

	utils.JSON(w, http.StatusOK,"Password updated successfully", data)
}