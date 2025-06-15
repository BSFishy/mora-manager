package wingman

import "context"

type contextKey string

const moduleKey contextKey = "module"

func withModule(ctx context.Context, module string) context.Context {
	return context.WithValue(ctx, moduleKey, module)
}

func ModuleName(ctx context.Context) string {
	name, _ := ctx.Value(moduleKey).(string)
	return name
}
