package routes

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/gorm"

	// _ "github.com/thoraf20/loanee/docs"
	"github.com/thoraf20/loanee/config"
	utils "github.com/thoraf20/loanee/internal/utils"
)

func NewRouter(cfg *config.Config, db *sql.DB) http.Handler {
	r := mux.NewRouter()

	// API grouping
	api := r.PathPrefix("/api/v1").Subrouter()

	//auth
	authRouter := api.PathPrefix("/auth").Subrouter()
	HandleAuthRoutes(authRouter, &gorm.DB{})

	//user
	userRouter := api.PathPrefix("/user").Subrouter()
	HandleUserRoutes(userRouter, &gorm.DB{})

	//wallet
	walletRouter := api.PathPrefix("/wallets").Subrouter()
	HandleWalletRoutes(walletRouter, &gorm.DB{})

	//collateral
	collateralRouter := api.PathPrefix("/collaterals").Subrouter()
	HandleCollateralRoutes(collateralRouter, &gorm.DB{})

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		utils.JSON(w, http.StatusOK, "Service is healthy", map[string]interface{}{
				"status":      "ok",
				"timestamp":   time.Now().Format(time.RFC3339),
				"environment": cfg.AppEnv,
		})
	}).Methods(http.MethodGet)

	// Swagger JSON (optional if you didn’t generate docs.go)
	r.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "docs/swagger.json")
	}).Methods(http.MethodGet)

	// Swagger UI
	r.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"),
	))
	err := r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		t, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
		fmt.Printf("Registered route: %s %v\n", t, methods)
		return nil
	})
	if err != nil {
		fmt.Printf("error walking routes: %v\n", err)
	}

	return r
}