package routes

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/thoraf20/loanee/api/handlers"
	"github.com/thoraf20/loanee/internal/services"
	"github.com/thoraf20/loanee/internal/repo"

	"github.com/gorilla/mux"
)

func HandleAuthRoutes(api *mux.Router, db *sqlx.DB) {
	userRepo := repository.NewUserRepository(db)
	authService := services.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService)

	api.HandleFunc("/register", authHandler.RegisterUser).Methods(http.MethodPost)
	api.HandleFunc("/login", authHandler.LoginUser).Methods(http.MethodPost)
}