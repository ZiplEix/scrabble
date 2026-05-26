package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

type testCtxKeyType struct{}
var testCtxKey = testCtxKeyType{}

func TestHumanHandler(t *testing.T) {
	RegisterContextKey(testCtxKey, "request_id")

	var buf bytes.Buffer
	h := NewHumanHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})

	ctx := context.Background()
	ctx = context.WithValue(ctx, testCtxKey, "test-req-id")

	// Trigger Handle directly
	r := slog.NewRecord(time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC), slog.LevelInfo, "Hello World", 0)
	r.AddAttrs(slog.String("key1", "val1"))

	err := h.Handle(ctx, r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Hello World") {
		t.Errorf("expected output to contain 'Hello World', got %q", out)
	}
	if !strings.Contains(out, "key1=") || !strings.Contains(out, "val1") {
		t.Errorf("expected output to contain key1 and val1, got %q", out)
	}
}
