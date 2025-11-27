package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/thoraf20/loanee/config"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Build database URL
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)

	// Migration path (ensure relative path is correct)
	migrationDir := "file://db/migrations"

	// Get command argument: up, down, force, version, etc.
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run cmd/migrate/main.go [up|down|force|version]")
	}
	command := os.Args[1]

	m, err := migrate.New(migrationDir, dbURL)
	if err != nil {
		log.Fatalf("Failed to init migrate: %v", err)
	}

	switch command {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration up failed: %v", err)
		}
		fmt.Println("Migrations applied successfully")

	case "down":
		if err := m.Down(); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
		fmt.Println("Migrations rolled back successfully")

	case "version":
		v, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("Error getting version: %v", err)
		}
		fmt.Printf("Current migration version: %d (dirty: %v)\n", v, dirty)

	case "force":
		if len(os.Args) < 3 {
			log.Fatal("Usage: go run cmd/migrate/main.go force [version]")
		}
		version := os.Args[2]
		v, err := strconv.Atoi(version)
		if err != nil {
			log.Fatalf("Invalid version number: %v", err)
		}
		if err := m.Force(v); err != nil {
			log.Fatalf("Force failed: %v", err)
		}
		fmt.Printf("Forced migration version to %d\n", v)

	default:
		log.Fatalf("Unknown command: %s (expected up, down, version, force)", command)
	}
}
