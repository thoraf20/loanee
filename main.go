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
	// repository "github.com/thoraf20/loanee/internal/repo"
	// "github.com/thoraf20/loanee/internal/services"
	utils "github.com/thoraf20/loanee/internal/utils"
	// "github.com/thoraf20/loanee/pkg/chain"
)

func main() {
  if err := godotenv.Load(); err != nil {
		log.Warn().Msg("No .env file found, using environment variables")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configurations")
	}

	utils.InitLogger(cfg.AppEnv)
	log.Info().Msg("Logger initialized successfully")

	if err := database.Connect(cfg); 
	err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer closeDatabase()
	log.Info().Msg("Database connected successfully")

	database.RunMigrations(cfg)
	log.Info().Msg("Database migrations completed")

	goOrm, err := database.DB.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to obtain sql.DB from gorm.DB")
	}

	cache.InitRedis()
	log.Info().Msg("Redis cache initialized successfully")

	// collateralRepo := repository.NewCollateralRepository(database.DB)
	// priceService := services.NewPriceService()

	// ethVerifier := chain.NewEthereumVerifier(
	// 	cfg.Blockchain.Ethereum.RPCURL,
	// 	cfg.Blockchain.Ethereum.MinConfirmations,
	// )
	// ethVerifier, err := chain.NewEthereumVerifier(cfg.Blockchain.Ethereum.RPCURL, cfg.Blockchain.Ethereum.MinConfirmations)
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("failed to initialize Ethereum verifier")
	// }

	// verifiers := map[string]chain.ChainVerifier{
  //   "ETH": ethVerifier,
  //   "USDT": ethVerifier, // assuming ERC20 USDT for now
	// }

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

	waitForShutdown(server)
}

func closeDatabase() {
	sqlDB, err := database.DB.DB()
	if err != nil {
		log.Error().Err(err).Msg("Failed to obtain sql.DB for closing")
		return
	}
	if err := sqlDB.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close database connection")
	}
	log.Info().Msg("Database connection closed cleanly")
}

func waitForShutdown(server *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Warn().Msg("Server shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Forced to shutdown server")
	}

	log.Info().Msg("Server stopped cleanly")
}