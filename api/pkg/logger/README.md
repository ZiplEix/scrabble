# Logger Package

A self-contained, high-performance Go structured logging package built on top of the standard library's `slog`. It offers pretty terminal formatting for humans, JSON structured logging for machines, context-aware tracing attributes, and dynamic PostgreSQL asynchronous batch logging.

## Features

- **Human-Friendly Mode (`"human"`)**: Beautiful, high-contrast colored background level badges (`INFO`, `WARN`, `ERROR`, `DEBUG`) with perfect alignment, clean timestamps, and key-value logging.
- **Machine-Friendly Mode (`"machine"`)**: Standard structured JSON format optimized for log aggregators (e.g., Elasticsearch, Datadog).
- **Context Agnostic Key Extraction**: Decoupled from domain-specific context keys. You can dynamically register any context keys you want to automatically extract and log using `logger.RegisterContextKey(key, name)`.
- **PostgreSQL Async Batch Logging**: Features a robust database logging engine that aggregates log lines in-memory and periodically flushes them in batches using high-speed PostgreSQL copy protocols (`pq.CopyIn`).
- **Dynamic Database Logging Toggle**: Turn database persistence on or off at runtime using `logger.SaveToDB(bool)`.

---

## API Reference

### Initialization & Configuration

```go
// Init initializes the global slog logger.
// - mode: "human" (pretty terminal badges) or "machine" (JSON output)
// - db: optional *sql.DB database connection. If provided, activates PostgreSQL logging.
// Returns a shutdown function to gracefully flush pending database log batches.
func Init(mode string, db *sql.DB) func(context.Context) error
```

```go
// SaveToDB dynamically enables (true) or disables (false) database logging at runtime.
func SaveToDB(enabled bool)
```

### Context Registration Helpers

```go
// RegisterContextKey configures the logger to automatically extract a value from context under a specific key 
// and append it as a log attribute with the specified name.
func RegisterContextKey(key any, name string)
```

### Global Logging Functions

```go
func Info(ctx context.Context, msg string, args ...any)
func Warn(ctx context.Context, msg string, args ...any)
func Error(ctx context.Context, msg string, args ...any)
func Debug(ctx context.Context, msg string, args ...any)
```

---

## Usage Examples

### 1. Basic Setup & Lifecycle

```go
package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/ZiplEix/scrabble/api/pkg/logger"
	_ "github.com/lib/pq"
)

func main() {
	db, _ := sql.Open("postgres", "postgres://...")
	
	// Initialize logger in pretty "human" mode
	shutdown := logger.Init("human", db)
	
	// Enable DB logging dynamically
	logger.SaveToDB(true)

	// Ensure database logs are gracefully flushed before exit
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = shutdown(ctx)
	}()

	logger.Info(context.Background(), "Application started successfully")
}
```

### 2. Context-Agnostic Logging with Key Registration

Register any context keys in your application packages. The logger will automatically extract them.

```go
package mymiddleware

import (
	"context"
	"net/http"
	"github.com/ZiplEix/scrabble/api/pkg/logger"
)

// 1. Define custom, unexported key types to prevent collision
type requestIDKeyType struct{}
var requestIDKey = requestIDKeyType{}

func init() {
	// 2. Register key to the logger at startup
	logger.RegisterContextKey(requestIDKey, "request_id")
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 3. Inject value into context using standard context.WithValue
		ctx := context.WithValue(r.Context(), requestIDKey, "req-xyz-123")
		
		// All downstream logs using this context will automatically append: request_id="req-xyz-123"
		logger.Info(ctx, "HTTP request received")
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
```

## Database Schema Requirement

To use the PostgreSQL logging feature, ensure the target database contains a `logs` table compatible with the following schema:

```sql
CREATE TABLE IF NOT EXISTS logs (
    id SERIAL PRIMARY KEY,
    req_id TEXT,
    raw JSONB NOT NULL,
    received_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_logs_req_id ON logs(req_id);
```
