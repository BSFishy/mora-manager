package main

import (
	"context"

	"github.com/BSFishy/mora-manager/model"
	"github.com/BSFishy/mora-manager/state"
)

type contextKey string

const sessionKey contextKey = "session"

func WithSession(ctx context.Context, session *model.Session) context.Context {
	return context.WithValue(ctx, sessionKey, session)
}

func GetSession(ctx context.Context) (*model.Session, bool) {
	session, ok := ctx.Value(sessionKey).(*model.Session)
	return session, ok
}

const userKey contextKey = "user"

func WithUser(ctx context.Context, user *model.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func GetUser(ctx context.Context) (*model.User, bool) {
	user, ok := ctx.Value(userKey).(*model.User)
	return user, ok
}

const environmentKey contextKey = "environment"

func WithEnvironment(ctx context.Context, environment *model.Environment) context.Context {
	return context.WithValue(ctx, environmentKey, environment)
}

func GetEnvironment(ctx context.Context) (*model.Environment, bool) {
	environment, ok := ctx.Value(environmentKey).(*model.Environment)
	return environment, ok
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

const configKey contextKey = "config"

func WithConfig(ctx context.Context, config *Config) context.Context {
	return context.WithValue(ctx, configKey, config)
}

func GetConfig(ctx context.Context) (*Config, bool) {
	config, ok := ctx.Value(configKey).(*Config)
	return config, ok
}

const stateKey contextKey = "state"

func WithState(ctx context.Context, state *state.State) context.Context {
	return context.WithValue(ctx, stateKey, state)
}

func GetState(ctx context.Context) (*state.State, bool) {
	state, ok := ctx.Value(stateKey).(*state.State)
	return state, ok
}
