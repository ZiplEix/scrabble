package main

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/ZiplEix/scrabble/api/pkg/logger"
)

// StartLogRetention launches a background job that periodically deletes
// log rows older than maxAge from the `logs` table. Returns a stop function
// that gracefully stops the job.
func StartLogRetention(db *sql.DB, maxAge time.Duration, interval time.Duration) func(context.Context) error {
	if maxAge <= 0 {
		maxAge = 7 * 24 * time.Hour
	}
	if interval <= 0 {
		interval = 24 * time.Hour
	}

	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Run once at start
		purge(db, maxAge)

		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				purge(db, maxAge)
			}
		}
	}()

	return func(ctx context.Context) error {
		close(stop)
		done := make(chan struct{})
		go func() { wg.Wait(); close(done) }()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-done:
			return nil
		}
	}
}

func purge(db *sql.DB, maxAge time.Duration) {
	cutoff := time.Now().Add(-maxAge)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	res, err := db.ExecContext(ctx, `DELETE FROM logs WHERE received_at < $1`, cutoff)
	if err != nil {
		logger.Warn(ctx, "logs_retention_error", "error", err)
		return
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		logger.Info(ctx, "logs_retention_purge", "deleted", n)
	}
}
