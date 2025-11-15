package routes

import (
	"net/http"

	"github.com/thoraf20/loanee/api/handlers"
	database "github.com/thoraf20/loanee/db"
	"github.com/thoraf20/loanee/internal/repo"
	"github.com/thoraf20/loanee/internal/services"
	"gorm.io/gorm"

	"github.com/gorilla/mux"
)

func HandleAuthRoutes(api *mux.Router, db *gorm.DB) {
	if db == nil {
		panic("nil *gorm.DB passed to HandleAuthRoutes")
	}

	userRepo := repository.NewUserRepository(database.DB)
	authService := services.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService)

	api.HandleFunc("/register", authHandler.RegisterUser).Methods(http.MethodPost)
	api.HandleFunc("/verify-email", authHandler.VerifyEmail).Methods(http.MethodPost)
	api.HandleFunc("/login", authHandler.LoginUser).Methods(http.MethodPost)
	api.HandleFunc("/password-reset/request", authHandler.PasswordResetRequest).Methods(http.MethodPost)
	api.HandleFunc("/password/reset", authHandler.PasswordReset).Methods(http.MethodPost)
	api.HandleFunc("/resend-code", authHandler.ResendCode).Methods(http.MethodPost)
}