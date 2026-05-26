package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
)

// HumanHandler is a custom slog.Handler that outputs pretty, colored logs.
type HumanHandler struct {
	mu    *sync.Mutex
	w     io.Writer
	opts  slog.HandlerOptions
	attrs []slog.Attr
	group string
}

// NewHumanHandler creates a new HumanHandler.
func NewHumanHandler(w io.Writer, opts *slog.HandlerOptions) *HumanHandler {
	h := &HumanHandler{
		mu: &sync.Mutex{},
		w:  w,
	}
	if opts != nil {
		h.opts = *opts
	}
	return h
}

// Enabled checks if logging is enabled for the given level.
func (h *HumanHandler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

// Handle processes the log record.
func (h *HumanHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	const (
		Reset     = "\033[0m"
		Dim       = "\033[2m"
		Bold      = "\033[1m"
		BgDebug   = "\033[100;97m" // Grey background, White text
		BgInfo    = "\033[42;30m"  // Green background, Black text
		BgWarn    = "\033[43;30m"  // Yellow background, Black text
		BgError   = "\033[41;97m"  // Red background, White text
		BgDefault = "\033[45;97m"  // Magenta background, White text
	)

	var levelStr, levelBg string
	switch r.Level {
	case slog.LevelDebug:
		levelStr = " DEBUG "
		levelBg = BgDebug
	case slog.LevelInfo:
		levelStr = " INFO  "
		levelBg = BgInfo
	case slog.LevelWarn:
		levelStr = " WARN  "
		levelBg = BgWarn
	case slog.LevelError:
		levelStr = " ERROR "
		levelBg = BgError
	default:
		name := r.Level.String()
		if len(name) < 5 {
			levelStr = fmt.Sprintf(" %-5s ", name)
		} else if len(name) > 5 {
			levelStr = " " + name[:5] + " "
		} else {
			levelStr = " " + name + " "
		}
		levelBg = BgDefault
	}

	timeStr := r.Time.Format("15:04:05.000")

	// Print timestamp, level badge and base message
	fmt.Fprintf(h.w, "%s%s%s %s%s%s%s %s%s%s",
		Dim, timeStr, Reset,
		levelBg, Bold, levelStr, Reset,
		Bold, r.Message, Reset,
	)

	// Print pre-configured attributes
	for _, attr := range h.attrs {
		fmt.Fprintf(h.w, " %s%s=%s%v", Dim, attr.Key, Reset, attr.Value.Any())
	}

	// Print record attributes
	r.Attrs(func(attr slog.Attr) bool {
		fmt.Fprintf(h.w, " %s%s=%s%v", Dim, attr.Key, Reset, attr.Value.Any())
		return true
	})

	fmt.Fprintln(h.w)
	return nil
}

// WithAttrs returns a new HumanHandler with the given attributes.
func (h *HumanHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &HumanHandler{
		mu:    h.mu,
		w:     h.w,
		opts:  h.opts,
		attrs: newAttrs,
		group: h.group,
	}
}

// WithGroup returns a new HumanHandler with the given group.
func (h *HumanHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	return &HumanHandler{
		mu:    h.mu,
		w:     h.w,
		opts:  h.opts,
		attrs: h.attrs,
		group: name,
	}
}
