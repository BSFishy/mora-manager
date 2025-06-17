package expr

import "context"

type contextKey string

const functionRegistryKey contextKey = "function_registry"

func WithFunctionRegistry(ctx context.Context, registry *FunctionRegistry) context.Context {
	return context.WithValue(ctx, functionRegistryKey, registry)
}

func GetFunctionRegistry(ctx context.Context) (*FunctionRegistry, bool) {
	registry, ok := ctx.Value(functionRegistryKey).(*FunctionRegistry)
	return registry, ok
}
