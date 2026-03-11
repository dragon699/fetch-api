package telemetry

import (
	"log/slog"
	"os"
	"strings"
)


func NewLogger(logLevel string, logFormat string) *slog.Logger {
	level := slog.LevelInfo

	switch strings.ToLower(logLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	format := strings.ToLower(logFormat)
	var handler slog.Handler

	switch format {
	case "logfmt", "text", "":
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	default:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	return slog.New(handler)
}
