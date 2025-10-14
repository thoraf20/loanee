package database

import (
	"fmt"
	"log"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/thoraf20/loanee/config"
	"github.com/thoraf20/loanee/internal/models"
)

func RunMigrations(cfg *config.Config) {

	err := DB.AutoMigrate(
		&models.User{},
	)
	if err != nil {
		log.Fatalf("Migration init failed: %v", err)
	}

	fmt.Println("Database migrated successfully")
}