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

type pgxLogTracer struct{}

func (t *pgxLogTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	fmt.Println("Running query:", data.SQL, "Args:", data.Args)
	return ctx
}

func (t *pgxLogTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	fmt.Println("Query executed successfully:")
}

func Init(dsn string) error {
	var err error
	Pool, err = pgxpool.New(context.Background(), dsn)
	Pool.Config().ConnConfig.Tracer = &pgxLogTracer{}
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
	if Pool == nil {
		return nil, fmt.Errorf("database connection pool is not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	return rows, nil
}

func QueryRow(query string, args ...any) pgx.Row {
	if Pool == nil {
		log.Fatal("database connection pool is not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := Pool.QueryRow(ctx, query, args...)
	return row
}

func Exec(query string, args ...any) (pgconn.CommandTag, error) {
	log.Printf("Exec query: %s, args: %v\n", query, args)

	if Pool == nil {
		return pgconn.CommandTag{}, fmt.Errorf("database connection pool is not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan struct{})
	var err error

	go func() {
		_, err = Pool.Exec(ctx, query, args...)
		close(done)
	}()

	select {
	case <-done:
		if err != nil {
			log.Printf("Exec error: %v", err)
			return pgconn.CommandTag{}, fmt.Errorf("failed to execute statement: %w", err)
		}
	case <-ctx.Done():
		log.Println("Context timeout hit BEFORE Exec returned!")
		return pgconn.CommandTag{}, fmt.Errorf("context timeout before exec: %w", ctx.Err())
	}

	log.Printf("Exec successfully executed query: %s", query)
	return pgconn.CommandTag{}, nil
}
