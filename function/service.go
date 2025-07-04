package function

import (
	"context"

	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/point"
	"github.com/BSFishy/mora-manager/value"
)

func evaluateServiceFunction(ctx context.Context, deps expr.EvaluationContext, args expr.Args) (value.Value, []point.Point, error) {
	moduleName, err := args.Identifier(ctx, deps, 0)
	if err != nil {
		return nil, nil, err
	}

	serviceName, err := args.Identifier(ctx, deps, 1)
	if err != nil {
		return nil, nil, err
	}

	return value.NewServiceReference(moduleName, serviceName), nil, nil
}
