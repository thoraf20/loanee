package database

import (
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/thoraf20/loanee/config"
)

func RunMigrations(cfg *config.Config) {
	m, err := migrate.New(
		"file://db/migrations",
		cfg.Database.GetDatabaseString(),
	)
	if err != nil {
		log.Fatalf("Migration init failed: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("Database migrated successfully")
}