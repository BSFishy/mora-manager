package main

import (
	"context"
	"fmt"

	"github.com/BSFishy/mora-manager/config"
	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/util"
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
		Evaluate: func(ctx context.Context, args expr.Args) (value.Value, []config.Point, error) {
			moduleName, err := args.Identifier(ctx, 0)
			if err != nil {
				return nil, nil, err
			}

			serviceName, err := args.Identifier(ctx, 1)
			if err != nil {
				return nil, nil, err
			}

			return value.NewServiceReference(moduleName, serviceName), nil, nil
		},
	})
}

func evaluateConfigFunction(ctx context.Context, args expr.Args) (value.Value, []config.Point, error) {
	moduleName, identifier, err := getConfigNames(ctx, args)
	if err != nil {
		return nil, nil, err
	}

	state := util.Has(GetState(ctx))
	stateConfig := state.FindConfig(moduleName, identifier)
	if stateConfig != nil {
		return value.NewString(stateConfig.Value), []config.Point{}, nil
	}

	cfg := util.Has(GetConfig(ctx))
	c := cfg.FindConfig(moduleName, identifier)
	if c == nil {
		return nil, nil, fmt.Errorf("invalid config reference: (config %s %s)", moduleName, identifier)
	}

	return nil, []config.Point{*c}, nil
}

func getConfigNames(ctx context.Context, args expr.Args) (string, string, error) {
	var (
		moduleName string
		identifier string
		err        error
	)

	if args.Len() == 2 {
		moduleName, err = args.Identifier(ctx, 0)
		if err != nil {
			return "", "", fmt.Errorf("evaluating module name identifier: %w", err)
		}

		identifier, err = args.Identifier(ctx, 1)
		if err != nil {
			return "", "", fmt.Errorf("evaluating identifier: %w", err)
		}
	} else {
		moduleName = util.Has(util.GetModuleName(ctx))
		identifier, err = args.Identifier(ctx, 0)
		if err != nil {
			return "", "", fmt.Errorf("evaluating identifier: %w", err)
		}
	}

	return moduleName, identifier, nil
}
