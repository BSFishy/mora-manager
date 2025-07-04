package function

import (
	"context"
	"fmt"

	"github.com/BSFishy/mora-manager/expr"
)

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
