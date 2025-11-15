package config

import (
	"fmt"
	"os"


	"github.com/joho/godotenv"
)

type BlockchainConfig struct {
	Ethereum struct {
		RPCURL           string `mapstructure:"rpc_url"`
		MinConfirmations int    `mapstructure:"min_confirmations"`
	} `mapstructure:"ethereum"`
}

type Config struct {
	Database  DatabaseConfig
	Server    ServerConfig
	JWTSecret string
	AppEnv       string
	DefaultLTV string
	Blockchain BlockchainConfig `mapstructure:"blockchain"`
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	Env      string
}

type ServerConfig struct {
	AppPort string
}

func (d *DatabaseConfig) GetDatabaseString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.Name,
		d.SSLMode,
	)
}

func LoadConfig() (*Config, error) {

	godotenv.Load()

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASS", "postgres"),
			Name:     getEnv("DB_NAME", "loanee"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Server: ServerConfig{
			AppPort: getEnv("APP_PORT", "8080"),
		},
		JWTSecret: getEnv("JWT_SECRET", "your-256-bit-secret"),
		AppEnv:       getEnv("APP_ENV", "development"),
		DefaultLTV: getEnv("DEFAULT_LTV", "0.65"),
	}, nil
}

func getEnv(key string, defaultValue string) string {

	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Global config instance
var config *Config

// GetConfig returns the application configuration
// It loads the configuration if it hasn't been loaded yet
func GetConfig() *Config {
	if config == nil {
		var err error
		config, err = LoadConfig()
		if err != nil {
			panic("Failed to load configuration: " + err.Error())
		}
	}
	return config
}