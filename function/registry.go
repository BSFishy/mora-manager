package function

import (
	"context"
	"fmt"

	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/point"
	"github.com/BSFishy/mora-manager/value"
)

type Registry struct {
	manager WingmanManager
	builtin map[string]expr.ExpressionFunction
}

func NewRegistry(deps HasWingmanManager) *Registry {
	return &Registry{
		manager: deps.GetWingmanManager(),
		builtin: map[string]expr.ExpressionFunction{
			"config": {
				MinArgs:  1,
				MaxArgs:  2,
				Evaluate: evaluateConfigFunction,
			},
			"service": {
				MinArgs:  2,
				MaxArgs:  2,
				Evaluate: evaluateServiceFunction,
			},
		},
	}
}

func (r *Registry) Evaluate(ctx context.Context, deps expr.EvaluationContext, name string, args expr.Args) (value.Value, []point.Point, error) {
	if fn, ok := r.builtin[name]; ok {
		if fn.IsInvalid(args) {
			return nil, nil, fmt.Errorf("invalid arguments for: %s", name)
		}

		return fn.Evaluate(ctx, deps, args)
	}

	val, points, err := r.manager.EvaluateFunction(ctx, deps, name, args)
	if err != nil {
		return nil, nil, fmt.Errorf("evaluating wingman function: %w", err)
	}

	if val == nil && len(points) == 0 {
		return nil, nil, fmt.Errorf("invalid function: %s", name)
	}

	return val, points, nil
}
