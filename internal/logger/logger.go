package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

// Format encodes the logs to be strings when output
type Format string

// Stores the format types as constants
const (
	FormatJSON Format = "json" // structured, one JSON object per line - good for audit trails / log aggregators 
	FormatText Format = "text" // clear-text key=value - good for local dev diagnostics
)

// Config to control how the engine is constructed
type Config struct {
	Level string // "DEBUG", "INFO", "WARN", "ERROR" (case-insensitive)
	Format Format // FormatJSON or FormatText
	Output io.Writer // where records are written; defaults to os.Stdout
	AddSource bool      // include file:line of the log call site
}

// levelVar is shared by every logger the engine hands out, which is what
// makes runtime level switching possible: mutate it once and every derived
// logger immediately honors the new level.
var levelVar = new(slog.LevelVar)

var (
	defaultLogger *slog.Logger
	mu sync.RWMutex
)

// ParseLevel converts a config string into a slog.Level. Unknown values
// fall back to INFO rather than erroring, so a bad env var can't crash startup.
func ParseLevel(s string) slog.Level {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// New builds a *slog.Logger from Config and also sets it as the package
// default (so package-level Info/Debug/Warn/Error helpers below work).
func New(cfg Config) *slog.Logger {
	// Checks if the output is nil
	if cfg.Output == nil {
		// Makes the config output standard output
		cfg.Output = os.Stdout
	}

	// Parses the level from the config
	levelVar.Set(ParseLevel(cfg.Level))

	// Handles the level and the source
	handlerOpts := &slog.HandlerOptions{
		Level:     levelVar,
		AddSource: cfg.AddSource,
	}

	// Creates the log handler
	var handler slog.Handler

	// Decides the log format
	switch cfg.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(cfg.Output, handlerOpts)
	default:
		handler = slog.NewTextHandler(cfg.Output, handlerOpts)
	}

	// Creates a new logger
	l := slog.New(handler)

	// Locks the mutex
	mu.Lock()

	// Sets the logger as the default
	defaultLogger = l

	// Unlocks the logger
	mu.Unlock()

	// Returns the logger
	return l
}

// SetLevel changes the active log level at runtime for every logger derived
// from New (e.g. wire this to a SIGHUP handler or an admin HTTP endpoint).
func SetLevel(level string) {
	levelVar.Set(ParseLevel(level))
}

// CurrentLevel returns the active level, useful for a /debug/level endpoint.
func CurrentLevel() slog.Level {
	return levelVar.Level()
}

// Default returns the package-level default logger, building a sane
// INFO/text logger to stdout if New was never called.
func Default() *slog.Logger {
	mu.RLock()
	l := defaultLogger
	mu.RUnlock()
	if l == nil {
		return New(Config{Level: "INFO", Format: FormatText})
	}
	return l
}

// With returns a child logger with fixed structured fields attached to
// every record it emits — handy for per-request or per-module context.
func With(args ...any) *slog.Logger {
	return Default().With(args...)
}

// Convenience wrappers so call sites don't need to import log/slog directly.

func Debug(ctx context.Context, msg string, args ...any) {
	Default().DebugContext(ctx, msg, args...)
}

func Info(ctx context.Context, msg string, args ...any) {
	Default().InfoContext(ctx, msg, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	Default().WarnContext(ctx, msg, args...)
}

func Error(ctx context.Context, msg string, args ...any) {
	Default().ErrorContext(ctx, msg, args...)
}