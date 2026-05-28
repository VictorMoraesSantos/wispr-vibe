package logger

import (
	"log/slog"
	"os"
)

var log *slog.Logger

func Init(level string) {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: lvl,
	})
	log = slog.New(handler)
	slog.SetDefault(log)
}

func Get() *slog.Logger {
	if log == nil {
		Init("info")
	}
	return log
}
