package routes

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/thoraf20/loanee/config"
	utils "github.com/thoraf20/loanee/internal/utils"
)

func NewRouter(cfg *config.Config, db *sql.DB) http.Handler {
	r := mux.NewRouter()

	// API grouping
	api := r.PathPrefix("/api/v1").Subrouter()

	//auth
	authRouter := api.PathPrefix("/auth").Subrouter()
	HandleAuthRoutes(authRouter, &sqlx.DB{})

	//user
	// userRouter := api.PathPrefix("/users").Subrouter()
	// HandleUserRoutes(userRouter, &zerolog.Logger{})

	//loan
	// loan := api.PathPrefix("/loans").Subrouter()
	// loan.HandleFunc("/apply", loanHandler.ApplyLoan).Methods(http.MethodPost)
	// loan.HandleFunc("/{id}/repay", loanHandler.RepayLoan).Methods(http.MethodPost)
	// loan.HandleFunc("/{id}", loanHandler.GetLoan).Methods(http.MethodGet)

	// //wallet
	// wallet := api.PathPrefix("/wallets").Subrouter()
	// wallet.HandleFunc("/generate", walletHandler.GenerateAddress).Methods(http.MethodPost)
	// wallet.HandleFunc("/{userID}", walletHandler.GetWallets).Methods(http.MethodGet)

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		utils.JSON(w, http.StatusOK, "Service is healthy", map[string]interface{}{
			"status":     "ok",
			"timestamp":  time.Now().Format(time.RFC3339),
			"environment": cfg.AppEnv,
		})
	}).Methods(http.MethodGet)

	r.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // The URL pointing to API definition
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))

	return r
}