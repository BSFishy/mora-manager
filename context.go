package main

import (
	"context"

	"github.com/BSFishy/mora-manager/state"
	"k8s.io/client-go/kubernetes"
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

const clientsetKey contextKey = "clientset"

func WithClientset(ctx context.Context, client *kubernetes.Clientset) context.Context {
	return context.WithValue(ctx, clientsetKey, client)
}

func GetClientset(ctx context.Context) (*kubernetes.Clientset, bool) {
	client, ok := ctx.Value(clientsetKey).(*kubernetes.Clientset)
	return client, ok
}
