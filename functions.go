package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/BSFishy/mora-manager/util"
)

func RegisterConfigFunction(r *FunctionRegistry) {
	r.Register("config", ExpressionFunction{
		MinArgs:         1,
		MaxArgs:         2,
		Evaluate:        evaluateConfigFunction,
		GetConfigPoints: getConfigPointsConfigFunction,
	})
}

func evaluateConfigFunction(ctx context.Context, args Args) (*ReturnType, error) {
	state := util.Has(GetState(ctx))

	moduleName, identifier, err := getConfigNames(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("getting config names: %w", err)
	}

	if config := state.FindConfig(moduleName, identifier); config != nil {
		var (
			value string
			ok    bool
		)

		if value, ok = config.Value.(string); !ok {
			panic("config string is not a string")
		}

		returnValue := NewString(value)
		return &returnValue, nil
	}

	return nil, errors.New("invalid expression")
}

func getConfigPointsConfigFunction(ctx context.Context, args Args) ([]ConfigPoint, error) {
	config := util.Has(GetConfig(ctx))
	state := util.Has(GetState(ctx))

	moduleName, identifier, err := getConfigNames(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("getting config names: %w", err)
	}

	if config := state.FindConfig(moduleName, identifier); config != nil {
		return []ConfigPoint{}, nil
	}

	cfg := config.FindConfig(moduleName, identifier)
	if cfg == nil {
		return nil, errors.New("invalid config name")
	}

	name, err := cfg.Name.EvaluateString(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting config name: %w", err)
	}

	var description *string
	if cfg.Description != nil {
		desc, err := cfg.Description.EvaluateString(ctx)
		if err != nil {
			return nil, fmt.Errorf("getting description: %w", err)
		}

		description = &desc
	}

	return []ConfigPoint{
		{
			ModuleName:  moduleName,
			Identifier:  identifier,
			Name:        name,
			Description: description,
		},
	}, nil
}

func getConfigNames(ctx context.Context, args Args) (string, string, error) {
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
		moduleName = util.Has(GetModuleName(ctx))
		identifier, err = args.Identifier(ctx, 0)
		if err != nil {
			return "", "", fmt.Errorf("evaluating identifier: %w", err)
		}
	}

	return moduleName, identifier, nil
}
