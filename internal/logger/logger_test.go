package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  slog.Level
	}{
		{"DEBUG", slog.LevelDebug},
		{"debug", slog.LevelDebug},
		{"WARN", slog.LevelWarn},
		{"warn", slog.LevelWarn},
		{"WARNING", slog.LevelWarn},
		{"ERROR", slog.LevelError},
		{"error", slog.LevelError},
		{"INFO", slog.LevelInfo},
		{"info", slog.LevelInfo},
		// Unknown values fall back to INFO
		{"", slog.LevelInfo},
		{"TRACE", slog.LevelInfo},
		{"garbage", slog.LevelInfo},
		{"  DEBUG  ", slog.LevelDebug}, // trimmed
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseLevel(tt.input)
			if got != tt.want {
				t.Errorf("ParseLevel(%q) = %v; want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNew_WritesToOutput(t *testing.T) {
	var buf bytes.Buffer
	l := New(Config{
		Level:  "INFO",
		Format: FormatText,
		Output: &buf,
	})

	if l == nil {
		t.Fatal("New() returned nil logger")
	}

	l.Info("hello from test")

	output := buf.String()
	if !strings.Contains(output, "hello from test") {
		t.Errorf("expected log output to contain %q; got: %s", "hello from test", output)
	}
}

func TestNew_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	l := New(Config{
		Level:  "INFO",
		Format: FormatJSON,
		Output: &buf,
	})

	l.Info("json test message")

	output := buf.String()
	if !strings.Contains(output, `"msg"`) {
		t.Errorf("expected JSON output to contain \"msg\"; got: %s", output)
	}
	if !strings.Contains(output, "json test message") {
		t.Errorf("expected output to contain the message; got: %s", output)
	}
}

func TestNew_NilOutputDefaultsToStdout(t *testing.T) {
	// This should not panic when Output is nil.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("New() panicked with nil Output: %v", r)
		}
	}()
	l := New(Config{Level: "INFO", Format: FormatText, Output: nil})
	if l == nil {
		t.Fatal("New() returned nil logger")
	}
}

func TestSetLevel_ChangesActiveLevel(t *testing.T) {
	var buf bytes.Buffer
	New(Config{Level: "ERROR", Format: FormatText, Output: &buf})

	// At ERROR level, DEBUG/INFO/WARN should be suppressed.
	SetLevel("DEBUG")
	if CurrentLevel() != slog.LevelDebug {
		t.Errorf("CurrentLevel() = %v after SetLevel(DEBUG); want %v", CurrentLevel(), slog.LevelDebug)
	}

	SetLevel("ERROR")
	if CurrentLevel() != slog.LevelError {
		t.Errorf("CurrentLevel() = %v after SetLevel(ERROR); want %v", CurrentLevel(), slog.LevelError)
	}
}

func TestCurrentLevel_ReturnsActiveLevel(t *testing.T) {
	var buf bytes.Buffer
	New(Config{Level: "WARN", Format: FormatText, Output: &buf})

	if got := CurrentLevel(); got != slog.LevelWarn {
		t.Errorf("CurrentLevel() = %v; want %v", got, slog.LevelWarn)
	}
}

func TestDefault_ReturnsNonNil(t *testing.T) {
	// Reset defaultLogger to nil by re-initialising package state via a fresh call.
	// Default() must return a valid logger even if New() was never called.
	l := Default()
	if l == nil {
		t.Fatal("Default() returned nil")
	}
}

func TestWith_ReturnsChildLogger(t *testing.T) {
	var buf bytes.Buffer
	New(Config{Level: "INFO", Format: FormatText, Output: &buf})

	child := With("component", "test")
	if child == nil {
		t.Fatal("With() returned nil")
	}
	child.Info("child log")

	if !strings.Contains(buf.String(), "component") {
		t.Errorf("expected With() child to include key; got: %s", buf.String())
	}
}

func TestConvenienceWrappers_DoNotPanic(t *testing.T) {
	var buf bytes.Buffer
	New(Config{Level: "DEBUG", Format: FormatText, Output: &buf})
	ctx := context.Background()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("convenience wrapper panicked: %v", r)
		}
	}()

	Debug(ctx, "debug msg")
	Info(ctx, "info msg")
	Warn(ctx, "warn msg")
	Error(ctx, "error msg")
}
