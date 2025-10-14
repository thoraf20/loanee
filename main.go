package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"

	"github.com/thoraf20/loanee/api/routes"
	config "github.com/thoraf20/loanee/config"
	database "github.com/thoraf20/loanee/db"
	"github.com/thoraf20/loanee/internal/cache"
	utils "github.com/thoraf20/loanee/internal/utils"
)

func main() {
	 // Load .env manually before viper
  _ = godotenv.Load()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	utils.InitLogger(cfg.AppEnv)
	log.Info().Msg("Logger initialized successfully")

	if err := database.Connect(cfg); 
	err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	defer func() {
		sqlDB, err := database.DB.DB()
		if err != nil {
			log.Error().Err(err).Msg("failed to obtain sql.DB to close")
			return
		}
		if cerr := sqlDB.Close(); 
		cerr != nil {
			log.Error().Err(cerr).Msg("failed to close database connection")
		}
	}()

	log.Info().Msg("Database connected successfully")

	database.RunMigrations(cfg)
	log.Info().Msg("Database migrations completed")

	// extract the underlying *sql.DB from gorm
	goOrm, err := database.DB.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to obtain sql.DB from gorm.DB")
	}

	cache.InitRedis()

	// routes.NewRouter expects *sql.DB, pass the underlying *sql.DB
	router := routes.NewRouter(cfg, goOrm)

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	server := &http.Server{
		Addr:              ":" + cfg.Server.AppPort,
		Handler:           corsMiddleware.Handler(router),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info().Msgf("Server started on port %s", cfg.Server.AppPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server crashed unexpectedly")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Warn().Msg("Server shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("forced to shutdown server")
	}

	log.Info().Msg("Server stopped cleanly")
}