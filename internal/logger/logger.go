package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type ctxKey int

const (
	envFormat        = "QUIZBATTLE_LOG_FORMAT"
	envLevel         = "QUIZBATTLE_LOG_LEVEL"
	logIDKey  ctxKey = iota
)

func New() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     parseLevel(os.Getenv(envLevel)),
		AddSource: true,
	}

	var handler slog.Handler
	switch strings.ToLower(os.Getenv(envFormat)) {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(handler)

}

func WithLogID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, logIDKey, id)
}

func FromContext(ctx context.Context, base *slog.Logger) *slog.Logger {
	if id, ok := ctx.Value(logIDKey).(string); ok && id != "" {
		return base.With("log_id", id)
	}
	return base
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "error":
		return slog.LevelError
	case "warn":
		return slog.LevelWarn
	default:
		return slog.LevelInfo

	}
}
