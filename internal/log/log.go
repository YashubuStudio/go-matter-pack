package log

import (
	"log/slog"
	"os"
)

// NewLogger returns a slog.Logger configured with the provided level.
func NewLogger(level slog.Leveler) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}
