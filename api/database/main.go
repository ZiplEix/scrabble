package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Init(dsn string) error {
	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open DB: %w", err)
	}

	// Configure la taille max des connexions (optionnel)
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// Retry ping avec timeout
	for i := range 5 {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err = DB.PingContext(ctx)
		cancel()
		if err == nil {
			log.Println("Connected to PostgreSQL successfully")
			return nil
		}
		log.Printf("Failed to connect to PostgreSQL, retrying in 2 seconds... (%d/5)", i+1)
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("failed to connect to PostgreSQL after retries: %w", err)
}

func Query(query string, args ...any) (*sql.Rows, error) {
	if DB == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	return rows, nil
}

func QueryRow(query string, args ...any) *sql.Row {
	if DB == nil {
		log.Fatal("database connection is not initialized")
	}

	return DB.QueryRow(query, args...)
}

func Exec(query string, args ...any) (sql.Result, error) {
	log.Printf("Exec query: %s, args: %v\n", query, args)
	if DB == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	result, err := DB.Exec(query, args...)
	if err != nil {
		log.Printf("Exec error: %v", err)
		return nil, fmt.Errorf("failed to execute statement: %w", err)
	}
	log.Printf("Exec successfully executed query: %s", query)
	return result, nil
}
