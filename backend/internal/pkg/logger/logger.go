package logger

import (
	"log/slog"
	"os"
)

// New returns a structured *slog.Logger.  production=true → JSON, else → Text (彩色友好).
func New(production bool) *slog.Logger {
	var handler slog.Handler
	if production {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}
	return slog.New(handler)
}
