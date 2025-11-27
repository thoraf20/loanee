package container

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/thoraf20/loanee/config"
	"github.com/thoraf20/loanee/internal/auth"
	"github.com/thoraf20/loanee/internal/blockchain"
	"github.com/thoraf20/loanee/internal/collateral"
	"github.com/thoraf20/loanee/internal/loan"
	"github.com/thoraf20/loanee/internal/models"
	"github.com/thoraf20/loanee/internal/payment"
	"github.com/thoraf20/loanee/internal/pricing"
	"github.com/thoraf20/loanee/internal/user"
	jwt "github.com/thoraf20/loanee/internal/utils"
	"github.com/thoraf20/loanee/internal/wallet"
	"github.com/thoraf20/loanee/pkg/tokenblacklist"

	"github.com/redis/go-redis/v9"
	"github.com/thoraf20/loanee/pkg/validator"
	"gorm.io/gorm"
)

// Container holds all application dependencies
type Container struct {
	Config *config.Config
	Logger zerolog.Logger
	DB     *gorm.DB

	// Validators
	Validator *validator.Validator

	// Repositories
	UserRepo       user.Repository
	AuthRepo       auth.Repository
	CollateralRepo collateral.Repository
	WalletRepo     wallet.Repository
	LoanRepo       loan.Repository
	PaymentRepo    payment.Repository

	// Services
	AuthService        *auth.Service
	UserService        *user.Service
	CollateralService  *collateral.Service
	WalletService      *wallet.Service
	LoanService        *loan.Service
	PaymentService     *payment.Service
	PricingService     pricing.Provider
	BlockchainVerifier blockchain.Verifier

	// Handlers
	AuthHandler       *auth.Handler
	UserHandler       *user.Handler
	CollateralHandler *collateral.Handler
	WalletHandler     *wallet.Handler
	LoanHandler       *loan.Handler
	PaymentHandler    *payment.Handler

	RedisClient    *redis.Client
	TokenBlacklist tokenblacklist.Blacklist
	JWTManager     *jwt.Manager
}

// New creates and initializes the dependency container
func New(cfg *config.Config, logger zerolog.Logger) (*Container, error) {
	c := &Container{
		Config: cfg,
		Logger: logger,
	}

	// Initialize in order of dependencies
	if err := c.initDatabase(); err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	if err := c.initRedis(); err != nil {
		return nil, fmt.Errorf("failed to init redis: %w", err)
	}

	if err := c.initTokenBlacklist(); err != nil {
		return nil, fmt.Errorf("failed to init token blacklist: %w", err)
	}

	if err := c.initValidator(); err != nil {
		return nil, fmt.Errorf("failed to init validator: %w", err)
	}

	if err := c.initJWTManager(); err != nil {
		return nil, fmt.Errorf("failed to init jwt manager: %w", err)
	}

	if err := c.initBlockchainVerifier(); err != nil {
		return nil, fmt.Errorf("failed to init blockchain verifier: %w", err)
	}

	if err := c.initRepositories(); err != nil {
		return nil, fmt.Errorf("failed to init repositories: %w", err)
	}

	if err := c.initServices(); err != nil {
		return nil, fmt.Errorf("failed to init services: %w", err)
	}

	if err := c.initHandlers(); err != nil {
		return nil, fmt.Errorf("failed to init handlers: %w", err)
	}

	c.Logger.Info().Msg("Container initialized successfully")
	return c, nil
}

func (c *Container) initRedis() error {
	if !c.Config.Redis.Enabled {
		c.Logger.Info().Msg("Redis disabled, using in-memory token blacklist")
		return nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     c.Config.Redis.Address(),
		Password: c.Config.Redis.Password,
		DB:       c.Config.Redis.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		c.Logger.Warn().Err(err).Msg("Failed to connect to Redis, falling back to in-memory blacklist")
		return nil
	}

	c.RedisClient = client
	c.Logger.Info().Msg("Redis connected successfully")
	return nil
}

func (c *Container) initTokenBlacklist() error {
	if c.RedisClient != nil {
		c.TokenBlacklist = tokenblacklist.NewRedisBlacklist(c.RedisClient, c.Logger)
		c.Logger.Info().Msg("Using Redis token blacklist")
	} else {
		c.TokenBlacklist = tokenblacklist.NewMemoryBlacklist(c.Logger)
		c.Logger.Info().Msg("Using in-memory token blacklist")
	}
	return nil
}

// initJWTManager initializes JWT manager
func (c *Container) initJWTManager() error {
	c.JWTManager = jwt.NewManager(c.Config)
	c.Logger.Info().Msg("JWT manager initialized")
	return nil
}

func (c *Container) initBlockchainVerifier() error {
	if c.Config.Blockchain.UseNoopVerifier || c.Config.Blockchain.EthereumRPC == "" {
		c.BlockchainVerifier = blockchain.NewNoopVerifier()
		c.Logger.Warn().Msg("Using noop blockchain verifier")
		return nil
	}

	verifier, err := blockchain.NewEthereumVerifier(
		c.Config.Blockchain.EthereumRPC,
		c.Config.Blockchain.MinConfirmations,
	)
	if err != nil {
		c.Logger.Error().Err(err).Msg("Failed to initialize Ethereum verifier, falling back to noop")
		c.BlockchainVerifier = blockchain.NewNoopVerifier()
		return nil
	}

	c.BlockchainVerifier = verifier
	c.Logger.Info().Msg("Blockchain verifier initialized")
	return nil
}

// initDatabase initializes GORM database connection
func (c *Container) initDatabase() error {
	db, err := config.NewDatabase(c.Config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run auto migrations
	if err := db.AutoMigrate(
		&user.User{},
		&user.VerificationCode{},
		&user.PasswordResetToken{},
		&models.Collateral{},
		&models.Wallet{},
		&models.Loan{},
		&models.Payment{},
	); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	c.DB = db
	c.Logger.Info().Msg("Database connected successfully")
	return nil
}

// initValidator initializes request validator
func (c *Container) initValidator() error {
	c.Validator = validator.New()
	c.Logger.Info().Msg("Validator initialized")
	return nil
}

// initRepositories initializes all repositories
func (c *Container) initRepositories() error {
	c.UserRepo = user.NewRepository(c.DB, c.Logger)
	c.AuthRepo = auth.NewRepository(c.DB, c.Logger)
	c.CollateralRepo = collateral.NewRepository(c.DB, c.Logger)
	c.WalletRepo = wallet.NewRepository(c.DB, c.Logger)
	c.LoanRepo = loan.NewRepository(c.DB, c.Logger)
	c.PaymentRepo = payment.NewRepository(c.DB, c.Logger)

	c.Logger.Info().Msg("Repositories initialized")
	return nil
}

// initServices initializes all services
func (c *Container) initServices() error {
	// Pricing service (external API)
	c.PricingService = pricing.NewCoinGeckoProvider(
		c.Config.CoinGecko.APIKey,
		c.Logger,
	)

	if c.BlockchainVerifier == nil {
		c.BlockchainVerifier = blockchain.NewNoopVerifier()
	}

	// User service
	c.UserService = user.NewService(
		c.UserRepo,
		c.Logger,
	)

	// Wallet service
	c.WalletService = wallet.NewService(
		c.WalletRepo,
		c.Logger,
	)

	// Loan service
	c.LoanService = loan.NewService(
		c.LoanRepo,
		c.Config,
		c.Logger,
	)

	// Payment service
	c.PaymentService = payment.NewService(
		c.PaymentRepo,
		c.LoanService,
		c.Logger,
	)

	// Auth service
	c.AuthService = auth.NewService(
		c.AuthRepo,
		c.JWTManager,
		c.TokenBlacklist,
		c.Config,
		c.Logger,
	)

	// Collateral service
	c.CollateralService = collateral.NewService(
		c.CollateralRepo,
		c.PricingService,
		c.BlockchainVerifier,
		c.LoanService,
		c.Config,
		c.Logger,
	)

	c.Logger.Info().Msg("Services initialized")
	return nil
}

// initHandlers initializes all HTTP handlers
func (c *Container) initHandlers() error {
	c.AuthHandler = auth.NewHandler(
		c.AuthService,
		c.Validator,
		c.Logger,
	)

	c.UserHandler = user.NewHandler(
		c.UserService,
		c.Validator,
		c.Logger,
	)

	c.CollateralHandler = collateral.NewHandler(
		c.CollateralService,
		c.Validator,
		c.Logger,
	)

	c.WalletHandler = wallet.NewHandler(
		c.WalletService,
		c.Logger,
	)

	c.LoanHandler = loan.NewHandler(
		c.LoanService,
		c.Logger,
	)

	c.PaymentHandler = payment.NewHandler(
		c.PaymentService,
		c.Logger,
	)

	c.Logger.Info().Msg("Handlers initialized")
	return nil
}

// Shutdown gracefully shuts down all resources
func (c *Container) Shutdown() error {
	c.Logger.Info().Msg("Shutting down container...")

	// Close Redis connection
	if c.RedisClient != nil {
		if err := c.RedisClient.Close(); err != nil {
			c.Logger.Error().Err(err).Msg("Failed to close Redis connection")
		}
	}

	// Close database
	sqlDB, err := c.DB.DB()
	if err == nil {
		sqlDB.Close()
	}

	return nil
}
