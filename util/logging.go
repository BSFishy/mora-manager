package util

import (
	"context"
	"log/slog"
)

type contextKey string

const loggerKey = contextKey("logger")

func LogFromCtx(ctx context.Context) *slog.Logger {
	l, ok := ctx.Value(loggerKey).(*slog.Logger)
	if !ok {
		return slog.Default()
	}

	return l
}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}
