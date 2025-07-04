package main

import (
	"context"
	"fmt"

	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/point"
	"github.com/BSFishy/mora-manager/value"
)

func RegisterDefaultFunctions(r *expr.FunctionRegistry) {
	r.Register("config", expr.ExpressionFunction{
		MinArgs:  1,
		MaxArgs:  2,
		Evaluate: evaluateConfigFunction,
	})

	r.Register("service", expr.ExpressionFunction{
		MinArgs: 2,
		MaxArgs: 2,
		Evaluate: func(ctx context.Context, deps expr.EvaluationContext, args expr.Args) (value.Value, []point.Point, error) {
			moduleName, err := args.Identifier(ctx, deps, 0)
			if err != nil {
				return nil, nil, err
			}

			serviceName, err := args.Identifier(ctx, deps, 1)
			if err != nil {
				return nil, nil, err
			}

			return value.NewServiceReference(moduleName, serviceName), nil, nil
		},
	})
}

func evaluateConfigFunction(ctx context.Context, deps expr.EvaluationContext, args expr.Args) (value.Value, []point.Point, error) {
	moduleName, identifier, err := getConfigNames(ctx, deps, args)
	if err != nil {
		return nil, nil, err
	}

	state := deps.GetState()
	stateConfig := state.FindConfig(moduleName, identifier)
	if stateConfig != nil {
		if stateConfig.Kind == point.String {
			return value.NewString(string(stateConfig.Value)), nil, nil
		} else if stateConfig.Kind == point.Secret {
			return value.NewSecret(string(stateConfig.Value)), nil, nil
		} else {
			return nil, nil, fmt.Errorf("invalid config kind: %s", stateConfig.Kind)
		}
	}

	cfg := deps.GetConfig()
	c := cfg.FindConfig(moduleName, identifier)
	if c == nil {
		return nil, nil, fmt.Errorf("invalid config reference: (config %s %s)", moduleName, identifier)
	}

	return nil, []point.Point{*c}, nil
}

func getConfigNames(ctx context.Context, deps expr.EvaluationContext, args expr.Args) (string, string, error) {
	var (
		moduleName string
		identifier string
		err        error
	)

	if args.Len() == 2 {
		moduleName, err = args.Identifier(ctx, deps, 0)
		if err != nil {
			return "", "", fmt.Errorf("evaluating module name identifier: %w", err)
		}

		identifier, err = args.Identifier(ctx, deps, 1)
		if err != nil {
			return "", "", fmt.Errorf("evaluating identifier: %w", err)
		}
	} else {
		moduleName = deps.GetModuleName()
		identifier, err = args.Identifier(ctx, deps, 0)
		if err != nil {
			return "", "", fmt.Errorf("evaluating identifier: %w", err)
		}
	}

	return moduleName, identifier, nil
}
