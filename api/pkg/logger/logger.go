package logger

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
)

type contextKeyMapping struct {
	key  any
	name string
}

var (
	globalLogger   *slog.Logger
	dbEnabled      atomic.Int32 // atomic boolean
	registeredKeys []contextKeyMapping
	keysMu         sync.RWMutex
)

// RegisterContextKey registers a context key to be automatically extracted.
func RegisterContextKey(key any, name string) {
	keysMu.Lock()
	defer keysMu.Unlock()
	registeredKeys = append(registeredKeys, contextKeyMapping{key: key, name: name})
}

// GetContextAttrs extracts all registered context attributes from context.
func GetContextAttrs(ctx context.Context) []slog.Attr {
	if ctx == nil {
		return nil
	}
	keysMu.RLock()
	defer keysMu.RUnlock()

	if len(registeredKeys) == 0 {
		return nil
	}

	attrs := make([]slog.Attr, 0, len(registeredKeys))
	for _, m := range registeredKeys {
		if val := ctx.Value(m.key); val != nil {
			attrs = append(attrs, slog.Any(m.name, val))
		}
	}
	return attrs
}



// Init initializes the global logger.
func Init(mode string, db *sql.DB) func(context.Context) error {
	var baseHandler slog.Handler

	if mode == "human" {
		baseHandler = NewHumanHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		baseHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	ctxHandler := &contextHandler{Handler: baseHandler}
	closeFn := func(ctx context.Context) error { return nil }

	if db != nil {
		pgH, pgClose := NewPGHandler(db, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
		closeFn = pgClose

		compositeHandler := &teeHandler{
			stdout: ctxHandler,
			db:     pgH,
		}
		globalLogger = slog.New(compositeHandler)
	} else {
		globalLogger = slog.New(ctxHandler)
	}

	slog.SetDefault(globalLogger)
	return closeFn
}

// SaveToDB enables or disables PostgreSQL logging.
func SaveToDB(enabled bool) {
	if enabled {
		dbEnabled.Store(1)
	} else {
		dbEnabled.Store(0)
	}
}

// IsDBEnabled checks if PostgreSQL logging is enabled.
func IsDBEnabled() bool {
	return dbEnabled.Load() == 1
}
type contextHandler struct {
	slog.Handler
}
func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs := GetContextAttrs(ctx); len(attrs) > 0 {
		r.AddAttrs(attrs...)
	}
	return h.Handler.Handle(ctx, r)
}

type teeHandler struct {
	stdout slog.Handler
	db     slog.Handler
}

func (t *teeHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return t.stdout.Enabled(ctx, l) || (IsDBEnabled() && t.db.Enabled(ctx, l))
}

func (t *teeHandler) Handle(ctx context.Context, r slog.Record) error {
	var err error
	if t.stdout.Enabled(ctx, r.Level) {
		err = t.stdout.Handle(ctx, r)
	}
	if IsDBEnabled() && t.db.Enabled(ctx, r.Level) {
		// Standard context enrichment for the DB logs too
		if attrs := GetContextAttrs(ctx); len(attrs) > 0 {
			r.AddAttrs(attrs...)
		}
		_ = t.db.Handle(ctx, r)
	}
	return err
}

func (t *teeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &teeHandler{
		stdout: t.stdout.WithAttrs(attrs),
		db:     t.db.WithAttrs(attrs),
	}
}

func (t *teeHandler) WithGroup(name string) slog.Handler {
	return &teeHandler{
		stdout: t.stdout.WithGroup(name),
		db:     t.db.WithGroup(name),
	}
}

// Global functions
func Info(ctx context.Context, msg string, args ...any) {
	if ctx == nil {
		ctx = context.Background()
	}
	slog.InfoContext(ctx, msg, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	if ctx == nil {
		ctx = context.Background()
	}
	slog.WarnContext(ctx, msg, args...)
}

func Error(ctx context.Context, msg string, args ...any) {
	if ctx == nil {
		ctx = context.Background()
	}
	slog.ErrorContext(ctx, msg, args...)
}

func Debug(ctx context.Context, msg string, args ...any) {
	if ctx == nil {
		ctx = context.Background()
	}
	slog.DebugContext(ctx, msg, args...)
}
