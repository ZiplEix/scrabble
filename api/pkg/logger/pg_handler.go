package logger

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"log/slog"
	"sync"
	"time"

	"github.com/lib/pq"
)

// PGHandler is a custom slog.Handler that writes logs to a PostgreSQL database table asynchronously in batches.
type PGHandler struct {
	db        *sql.DB
	opts      slog.HandlerOptions
	attrs     []slog.Attr
	group     string
	ch        chan []byte
	stop      chan struct{}
	wg        sync.WaitGroup
	batchSize int
	maxWait   time.Duration
}

// NewPGHandler creates a new PGHandler and a shutdown function.
func NewPGHandler(db *sql.DB, opts *slog.HandlerOptions) (*PGHandler, func(context.Context) error) {
	h := &PGHandler{
		db:        db,
		ch:        make(chan []byte, 10_000),
		stop:      make(chan struct{}),
		batchSize: 1000,
		maxWait:   2 * time.Second,
	}
	if opts != nil {
		h.opts = *opts
	}

	h.wg.Add(1)
	go h.loop()

	closeFn := func(ctx context.Context) error {
		close(h.stop)
		done := make(chan struct{})
		go func() {
			h.wg.Wait()
			close(done)
		}()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-done:
			return nil
		}
	}

	return h, closeFn
}

// Enabled checks if logging is enabled for the given level.
func (h *PGHandler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

// Handle serializes the record to JSON and pushes it to the buffer channel.
func (h *PGHandler) Handle(ctx context.Context, r slog.Record) error {
	logMap := make(map[string]any)
	logMap["ts"] = r.Time.Format(time.RFC3339Nano)
	logMap["level"] = r.Level.String()
	logMap["msg"] = r.Message

	// Add predefined attributes
	for _, attr := range h.attrs {
		logMap[attr.Key] = attr.Value.Any()
	}

	// Add record attributes
	r.Attrs(func(attr slog.Attr) bool {
		logMap[attr.Key] = attr.Value.Any()
		return true
	})

	// Add registered context attributes dynamically
	for _, attr := range GetContextAttrs(ctx) {
		logMap[attr.Key] = attr.Value.Any()
	}

	data, err := json.Marshal(logMap)
	if err != nil {
		return err
	}

	select {
	case h.ch <- data:
	default:
		log.Println("logger pg_handler: buffer full, dropping log")
	}

	return nil
}

// WithAttrs returns a new PGHandler with the given attributes.
func (h *PGHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &PGHandler{
		db:        h.db,
		opts:      h.opts,
		attrs:     newAttrs,
		group:     h.group,
		ch:        h.ch,
		stop:      h.stop,
		batchSize: h.batchSize,
		maxWait:   h.maxWait,
	}
}

// WithGroup returns a new PGHandler with the given group name.
func (h *PGHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	return &PGHandler{
		db:        h.db,
		opts:      h.opts,
		attrs:     h.attrs,
		group:     name,
		ch:        h.ch,
		stop:      h.stop,
		batchSize: h.batchSize,
		maxWait:   h.maxWait,
	}
}

func (h *PGHandler) loop() {
	defer h.wg.Done()
	ticker := time.NewTicker(h.maxWait)
	defer ticker.Stop()

	batch := make([][]byte, 0, h.batchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}

		tx, err := h.db.Begin()
		if err != nil {
			log.Printf("logger pg_handler: begin transaction error: %v", err)
			batch = batch[:0]
			return
		}

		stmt, err := tx.Prepare(pq.CopyIn("logs", "req_id", "raw"))
		if err != nil {
			log.Printf("logger pg_handler: prepare copy error: %v", err)
			_ = tx.Rollback()
			batch = batch[:0]
			return
		}

		for _, line := range batch {
			var tmp map[string]any
			if json.Unmarshal(line, &tmp) != nil {
				continue
			}

			var reqID string
			for _, k := range []string{"request_id", "req_id", "X-Request-ID", "X-Correlation-ID", "requestId"} {
				if v, ok := tmp[k]; ok {
					if s, ok := v.(string); ok && s != "" {
						reqID = s
						break
					}
				}
			}

			if _, err := stmt.Exec(reqID, string(line)); err != nil {
				log.Printf("logger pg_handler: exec copy error: %v", err)
			}
		}

		_, _ = stmt.Exec()
		_ = stmt.Close()
		if err := tx.Commit(); err != nil {
			log.Printf("logger pg_handler: commit error: %v", err)
		}
		batch = batch[:0]
	}

	for {
		select {
		case <-h.stop:
			flush()
			return
		case b := <-h.ch:
			batch = append(batch, b)
			if len(batch) >= h.batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}
