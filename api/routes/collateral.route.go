package routes

import (
	"github.com/rs/zerolog"
	"github.com/thoraf20/loanee/api/handlers"
	repository "github.com/thoraf20/loanee/internal/repo"
	"github.com/thoraf20/loanee/internal/services"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gorilla/mux"
)
func HandleCollateralRoutes(api *mux.Router, db *gorm.DB) {
	if db == nil {
		panic("nil *gorm.DB passed to HandleAuthRoutes")
	}

	logger, _ := zap.NewProduction()
	priceService := services.NewCoinGeckoProvider(logger, "")
	collateralRepo := repository.NewCollateralRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	collateralService := services.NewCollateralService(collateralRepo, priceService, &services.EthereumVerifier{}, walletRepo)
	collateralHandler := handlers.NewCollateralHandler(collateralService)

	protectedRouter := ApplyAuthMiddleware(api, &zerolog.Logger{})
	protectedRouter.HandleFunc("/preview", collateralHandler.PreviewCollateral).Methods("GET")
	protectedRouter.HandleFunc("/lock", collateralHandler.LockCollateral).Methods("POST")
	protectedRouter.HandleFunc("/create", collateralHandler.CreateCollateral).Methods("POST")
}
