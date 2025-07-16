package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Init(dsn string) error {
	var err error
	Pool, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := Pool.Ping(context.Background()); err != nil {
		for i := range 5 {
			log.Printf("Failed to connect to PostgreSQL, retrying in 2 seconds... (%d/5)", i+1)
			if err := Pool.Ping(context.Background()); err == nil {
				log.Println("Connected to PostgreSQL after retry")
				return nil
			}
			time.Sleep(2 * time.Second)
		}
		return fmt.Errorf("failed to connect to PostgreSQL after retries: %w", err)
	}
	log.Printf("Connected to PostgreSQL successfully")
	return nil
}

func Query(query string, args ...any) (pgx.Rows, error) {
	rows, err := Pool.Query(context.Background(), query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	return rows, nil
}

func QueryRow(query string, args ...any) pgx.Row {
	row := Pool.QueryRow(context.Background(), query, args...)
	return row
}

func Exec(query string, args ...any) (pgconn.CommandTag, error) {
	result, err := Pool.Exec(context.Background(), query, args...)
	if err != nil {
		return pgconn.CommandTag{}, fmt.Errorf("failed to execute statement: %w", err)
	}
	return result, nil
}
