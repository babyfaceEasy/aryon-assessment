package logger

import (
	"log/slog"
	"os"
)

// NewProductionLogger returns a logger configured to output JSON formatted logs.
func NewProductionLogger() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo, // Set the log level to info.
	})
	return slog.New(handler)
}
