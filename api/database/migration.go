package database

import (
	"embed"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

// RunMigrations exécute toutes les migrations Goose intégrées de façon automatique au démarrage de l'API.
func RunMigrations(embedFS embed.FS) error {
	_ = goose.SetDialect("postgres")

	// Configurer goose pour utiliser le système de fichiers embarqué
	goose.SetBaseFS(embedFS)

	log.Println("api: running embedded database migrations...")
	if err := goose.Up(DB, "migrations"); err != nil {
		return fmt.Errorf("failed to run goose migrations: %w", err)
	}

	log.Println("api: database migrations completed successfully")
	return nil
}
