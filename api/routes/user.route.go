package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/thoraf20/loanee/api/handlers"
	database "github.com/thoraf20/loanee/db"
	repository "github.com/thoraf20/loanee/internal/repo"
	"github.com/thoraf20/loanee/internal/services"
	"gorm.io/gorm"
)


func HandleUserRoutes(api *mux.Router, db *gorm.DB) {
	if db == nil {
		panic("nil *gorm.DB passed to HandleWalletRoutes")
	}

	userRepo := repository.NewUserRepository(database.DB)
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	protectedRouter := ApplyAuthMiddleware(api, &zerolog.Logger{})
	protectedRouter.HandleFunc("/me", userHandler.GetProfile).Methods(http.MethodGet)
}