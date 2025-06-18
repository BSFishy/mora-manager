package util

import (
	"context"
	"log/slog"
)

type contextKey string

const loggerKey contextKey = "logger"

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

const moduleNameKey contextKey = "module"

func WithModuleName(ctx context.Context, moduleName string) context.Context {
	return context.WithValue(ctx, moduleNameKey, moduleName)
}

func GetModuleName(ctx context.Context) (string, bool) {
	moduleName, ok := ctx.Value(moduleNameKey).(string)
	return moduleName, ok
}

const serviceNameKey contextKey = "service"

func WithServiceName(ctx context.Context, serviceName string) context.Context {
	return context.WithValue(ctx, serviceNameKey, serviceName)
}

func GetServiceName(ctx context.Context) (string, bool) {
	serviceName, ok := ctx.Value(serviceNameKey).(string)
	return serviceName, ok
}
