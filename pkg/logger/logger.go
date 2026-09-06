package logger

import (
	"context"
	"log/slog"
	"os"
)

var globalLogger *slog.Logger

func init() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	globalLogger = slog.New(handler)
	slog.SetDefault(globalLogger)
}

func Info(msg string, args ...any) {
	globalLogger.Info(msg, args...)
}

func Error(msg string, args ...any) {
	globalLogger.Error(msg, args...)
}

func Debug(msg string, args ...any) {
	globalLogger.Debug(msg, args...)
}

func Warn(msg string, args ...any) {
	globalLogger.Warn(msg, args...)
}

func With(args ...any) *slog.Logger {
	return globalLogger.With(args...)
}

func FromContext(ctx context.Context) *slog.Logger {
	return globalLogger
}
