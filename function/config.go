package function

import (
	"context"
	"fmt"

	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/point"
	"github.com/BSFishy/mora-manager/value"
)

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
