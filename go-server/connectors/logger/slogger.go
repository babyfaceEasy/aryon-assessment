package logger

import (
	"log/slog"
	"os"
	"strings"

	"connector-recruitment/go-server/connectors/config"
)

func getLogLevel(env config.Env) slog.Level {
	level := slog.LevelInfo // default

	switch strings.ToLower(env.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	return level
}

// NewProductionLogger returns a logger configured to output JSON formatted logs.
func NewProductionLogger(env config.Env) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: getLogLevel(env), // Set the log level to info.
	})
	newLogger := slog.New(handler).With("service", env.Name)
	return newLogger
}
