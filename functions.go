package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/BSFishy/mora-manager/util"
	"github.com/BSFishy/mora-manager/value"
)

func RegisterDefaultFunctions(r *FunctionRegistry) {
	r.Register("config", ExpressionFunction{
		MinArgs:  1,
		MaxArgs:  2,
		Evaluate: evaluateConfigFunction,
	})

	r.Register("service", ExpressionFunction{
		MinArgs: 2,
		MaxArgs: 2,
		Evaluate: func(ctx context.Context, args Args) (value.Value, []ConfigPoint, error) {
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

func evaluateConfigFunction(ctx context.Context, args Args) (value.Value, []ConfigPoint, error) {
	moduleName, identifier, err := getConfigNames(ctx, args)
	if err != nil {
		return nil, nil, err
	}

	state := util.Has(GetState(ctx))
	cfg := state.FindConfig(moduleName, identifier)
	if cfg != nil {
		return value.NewString(cfg.Value), []ConfigPoint{}, nil
	}

	config := util.Has(GetConfig(ctx))
	c := config.FindConfig(moduleName, identifier)
	if c == nil {
		return nil, nil, fmt.Errorf("invalid config reference: (config %s %s)", moduleName, identifier)
	}

	n, cfp, err := c.Name.Evaluate(ctx)
	if err != nil {
		return nil, nil, err
	}

	if len(cfp) > 0 {
		return value.NewNull(), cfp, nil
	}

	if n.Kind() != value.String {
		return nil, nil, errors.New("expected string")
	}

	var description *string
	if c.Description != nil {
		v, cfp, err := c.Description.Evaluate(ctx)
		if err != nil {
			return nil, nil, err
		}

		if len(cfp) > 0 {
			return value.NewNull(), cfp, nil
		}

		if v.Kind() != value.String {
			return nil, nil, errors.New("expected string")
		}

		str := v.String()
		description = &str
	}

	return value.NewNull(), []ConfigPoint{
		{
			ModuleName:  moduleName,
			Identifier:  identifier,
			Name:        n.String(),
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
