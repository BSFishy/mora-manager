package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

type contextKey string

const loggerKey = contextKey("logger")

func SetupLogger() {
	_, debug := os.LookupEnv("DEBUG")

	style, ok := os.LookupEnv("LOG_STYLE")
	if !ok {
		style = "text"
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		AddSource: debug,
	}

	switch style {
	case "text":
		handler = slog.NewTextHandler(os.Stderr, opts)
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		fmt.Fprintf(os.Stderr, "Unknown log style \"%s\", defaulting to text\n", style)
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	slog.SetDefault(slog.New(handler))
}

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
