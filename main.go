package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"syscall"
	"time"
	"os/signal"

	"github.com/rs/cors"
	"github.com/gorilla/mux"

	config "github.com/thoraf20/loanee/config"
	database "github.com/thoraf20/loanee/db"
	utils "github.com/thoraf20/loanee/internal/utils"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config")
	}

	utils.InitLogger(cfg.AppEnv)

	if err := database.Connect(cfg); 
	err != nil {
		panic("failed to connect to database")
	}

	err = database.DB.AutoMigrate()

	if err != nil {
		fmt.Printf("Migration error: %v\n", err)
		panic("failed to migrate database")
	}

	router := mux.NewRouter()

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	server := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: corsMiddleware.Handler(router),
	}

	// defer database.c

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				fmt.Printf("HTTP server error: %v\n", err)
			}
		}
	}()

	fmt.Println("Server started on port " + cfg.AppPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Server shutting down...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait until the timeout deadline
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %v\n", err)
	}

	fmt.Println("Server exited properly")
}