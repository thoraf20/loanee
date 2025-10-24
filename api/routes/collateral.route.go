package routes

import (
	"github.com/rs/zerolog"
	"github.com/thoraf20/loanee/api/handlers"
	repository "github.com/thoraf20/loanee/internal/repo"
	"github.com/thoraf20/loanee/internal/services"
	"gorm.io/gorm"

	"github.com/gorilla/mux"
)
func HandleCollateralRoutes(api *mux.Router, db *gorm.DB) {
	if db == nil {
		panic("nil *gorm.DB passed to HandleAuthRoutes")
	}

	priceService := services.NewPriceService()
	collateralRepo := repository.NewCollateralRepository(db)
	collateralService := services.NewCollateralService(collateralRepo, priceService)
	collateralHandler := handlers.NewCollateralHandler(collateralService)

	protectedRouter := ApplyAuthMiddleware(api, &zerolog.Logger{})
	protectedRouter.HandleFunc("/preview", collateralHandler.PreviewCollateral).Methods("GET")
	protectedRouter.HandleFunc("/lock", collateralHandler.LockCollateral).Methods("POST")
}
