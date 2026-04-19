package logger

import (
	"log/slog"
	"os"
	"strings"
)

// Init configures the global slog logger.
// logLevel: "debug", "info", "warn", "error" (default: "info")
// logFormat: "json" (default, for production), "text" (for development)
func Init(logLevel, logFormat string) {
	level := parseLevel(logLevel)

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug,
	}

	if strings.ToLower(logFormat) == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
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

// WithComponent returns a child logger with a "component" attribute.
func WithComponent(component string) *slog.Logger {
	return slog.Default().With("component", component)
}
