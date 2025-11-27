package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App        AppConfig
	Server     ServerConfig
	Database   DatabaseConfig
	JWT        JWTConfig
	Loan       LoanConfig
	CoinGecko  CoinGeckoConfig
	Log        LogConfig
	Redis      RedisConfig      `mapstructure:"redis"`
	Blockchain BlockchainConfig `mapstructure:"blockchain"`
}

type AppConfig struct {
	Name                     string `mapstructure:"name"`
	Environment              string `mapstructure:"environment"`
	Version                  string `mapstructure:"version"`
	RequireEmailVerification bool   `mapstructure:"require_email_verification"`
}

type ServerConfig struct {
	Port           int           `mapstructure:"port"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	IdleTimeout    time.Duration `mapstructure:"idle_timeout"`
	AllowedOrigins string        `mapstructure:"allowed_origins"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type JWTConfig struct {
	Secret             string        `mapstructure:"secret"`
	AccessTokenExpiry  time.Duration `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry time.Duration `mapstructure:"refresh_token_expiry"`
}

type LoanConfig struct {
	DefaultLTV             float64 `mapstructure:"default_ltv"`
	MaxLTV                 float64 `mapstructure:"max_ltv"`
	MinLTV                 float64 `mapstructure:"min_ltv"`
	DefaultInterestRate    float64 `mapstructure:"default_interest_rate"`
	LatePenaltyPerDay      float64 `mapstructure:"late_penalty_per_day"`
	RepaymentFrequencyDays int     `mapstructure:"repayment_frequency_days"`
	GracePeriodDays        int     `mapstructure:"grace_period_days"`
	PenaltyAPR             float64 `mapstructure:"penalty_apr"`
}

type RedisConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func (c *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type CoinGeckoConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

type BlockchainConfig struct {
	EthereumRPC      string `mapstructure:"ethereum_rpc"`
	MinConfirmations int    `mapstructure:"min_confirmations"`
	UseNoopVerifier  bool   `mapstructure:"use_noop_verifier"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/loanee/")

	// Set defaults
	setDefaults()

	// Read from environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("LOANEE")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		// Config file not found, using defaults and env vars
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func setDefaults() {
	// App defaults
	viper.SetDefault("app.name", "loanee")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("app.require_email_verification", true)

	// Server defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", 10*time.Second)
	viper.SetDefault("server.write_timeout", 10*time.Second)
	viper.SetDefault("server.idle_timeout", 120*time.Second)
	viper.SetDefault("server.allowed_origins", "*")

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "loanee")
	viper.SetDefault("database.sslmode", "disable")

	// Redis defaults
	viper.SetDefault("redis.enabled", false)
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// JWT defaults
	viper.SetDefault("jwt.secret", "your-secret-key-change-in-production")
	viper.SetDefault("jwt.access_token_expiry", 15*time.Minute)
	viper.SetDefault("jwt.refresh_token_expiry", 7*24*time.Hour)

	// Loan defaults
	viper.SetDefault("loan.default_ltv", 0.7)
	viper.SetDefault("loan.max_ltv", 0.8)
	viper.SetDefault("loan.min_ltv", 0.5)
	viper.SetDefault("loan.default_interest_rate", 8.5)
	viper.SetDefault("loan.late_penalty_per_day", 10.0)
	viper.SetDefault("loan.repayment_frequency_days", 30)
	viper.SetDefault("loan.grace_period_days", 3)
	viper.SetDefault("loan.penalty_apr", 15.0)

	// CoinGecko defaults
	viper.SetDefault("coingecko.base_url", "https://api.coingecko.com/api/v3")

	// Log defaults
	viper.SetDefault("log.level", "info")

	// Blockchain defaults
	viper.SetDefault("blockchain.ethereum_rpc", "")
	viper.SetDefault("blockchain.min_confirmations", 3)
	viper.SetDefault("blockchain.use_noop_verifier", true)
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.DBName,
		c.SSLMode,
	)
}
