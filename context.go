package main

import (
	"context"

	"github.com/BSFishy/mora-manager/state"
)

type contextKey string

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
