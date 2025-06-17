package wingman

import (
	"context"

	"github.com/BSFishy/mora-manager/state"
)

type contextKey string

const moduleKey contextKey = "module"

func withModule(ctx context.Context, module string) context.Context {
	return context.WithValue(ctx, moduleKey, module)
}

func ModuleName(ctx context.Context) string {
	name, _ := ctx.Value(moduleKey).(string)
	return name
}

const stateKey contextKey = "state"

func withState(ctx context.Context, state state.State) context.Context {
	return context.WithValue(ctx, stateKey, state)
}

func GetState(ctx context.Context) state.State {
	state, _ := ctx.Value(stateKey).(state.State)
	return state
}
