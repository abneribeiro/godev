// Package logging provides structured logging utilities for the godev application
package logging

import (
	"io"
	"log/slog"
	"os"
)

// Level represents logging levels
type Level = slog.Level

const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

// Config holds logging configuration
type Config struct {
	Level  Level
	Format string // "json" or "text"
	Output io.Writer
}

// DefaultConfig returns a default logging configuration
func DefaultConfig() *Config {
	return &Config{
		Level:  LevelInfo,
		Format: "text",
		Output: os.Stderr,
	}
}

// Setup initializes the global logger with the provided configuration
func Setup(config *Config) {
	if config == nil {
		config = DefaultConfig()
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: config.Level,
	}

	switch config.Format {
	case "json":
		handler = slog.NewJSONHandler(config.Output, opts)
	default:
		handler = slog.NewTextHandler(config.Output, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// GetLogger returns the default logger
func GetLogger() *slog.Logger {
	return slog.Default()
}

// WithContext returns a logger with context fields
func WithContext(fields ...interface{}) *slog.Logger {
	return slog.Default().With(fields...)
}
