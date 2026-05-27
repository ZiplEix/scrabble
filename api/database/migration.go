package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

// RunMigrations exécute toutes les migrations Goose de façon automatique au démarrage de l'API.
func RunMigrations() error {
	_ = goose.SetDialect("postgres")

	// Recherche du dossier migrations en remontant l'arborescence
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working dir: %w", err)
	}

	dir := wd
	migrationsDir := ""
	for i := 0; i <= 4; i++ {
		candidate := filepath.Join(dir, "migrations")
		if _, err := os.Stat(candidate); err == nil {
			migrationsDir = candidate
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	if migrationsDir == "" {
		return fmt.Errorf("could not find migrations directory upwards from %s", wd)
	}

	log.Printf("api: running database migrations from %s ...", migrationsDir)
	if err := goose.Up(DB, migrationsDir); err != nil {
		return fmt.Errorf("failed to run goose migrations: %w", err)
	}

	log.Println("api: database migrations completed successfully")
	return nil
}
