package main

import (
	"fmt"
	"log"
	"net/http"

	dbConfig "github.com/thoraf20/loanee/internal/config"
	database "github.com/thoraf20/loanee/internal/db"
	router "github.com/thoraf20/loanee/internal/http"
	utils "github.com/thoraf20/loanee/internal/utils"
)

func main() {
	cfg, err := dbConfig.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	utils.InitLogger(cfg.AppEnv)

	if err := database.Connect(cfg); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	router := router.NewRouter()

	addr := fmt.Sprintf(":%s", cfg.AppPort)
	fmt.Printf("[%s] Server running on http://localhost%s\n", cfg.AppEnv, addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}