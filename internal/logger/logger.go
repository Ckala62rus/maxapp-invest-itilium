// Package logger creates the shared application logger.
package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/config"
)

// New creates a configured logger for stdout and container log collection.
func New(cfg config.LoggingConfig) *slog.Logger {
	level := new(slog.LevelVar)
	level.Set(parseLevel(cfg.Level))

	options := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}

	if strings.EqualFold(cfg.Format, "text") {
		return slog.New(slog.NewTextHandler(os.Stdout, options))
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, options))
}

// parseLevel converts a string log level into the corresponding slog level.
func parseLevel(value string) slog.Level {
	switch strings.ToLower(value) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
