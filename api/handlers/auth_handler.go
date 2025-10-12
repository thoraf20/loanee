package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/thoraf20/loanee/internal/dtos"
	"github.com/thoraf20/loanee/internal/services"
	"github.com/thoraf20/loanee/internal/utils"
	"github.com/rs/zerolog/log"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RegisterUser handles user registration
func (h *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var input dtos.RegisterUserDTO

	if err := json.NewDecoder(r.Body).Decode(&input); 
	err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	user, err := h.authService.RegisterUser(ctx, input)
	if err != nil {
		log.Error().Err(err).Msg("failed to register user")
		utils.Error(w, http.StatusBadRequest, "Registration failed", err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, "User registered successfully", user)
}

// LoginUser handles user login and JWT generation
func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var input dtos.LoginDTO

	if err := json.NewDecoder(r.Body).Decode(&input); 
	err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	token, err := h.authService.LoginUser(ctx, input)
	if err != nil {
		log.Warn().Err(err).Msg("failed login attempt")
		utils.Error(w, http.StatusUnauthorized, "Invalid credentials", err.Error())
		return
	}

	response := map[string]string{"token": token}
	utils.JSON(w, http.StatusOK, "Login successful", response)
}