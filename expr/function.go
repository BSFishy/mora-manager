package expr

import (
	"context"
	"errors"

	"github.com/BSFishy/mora-manager/point"
	"github.com/BSFishy/mora-manager/value"
)

type ExpressionFunction struct {
	MinArgs  int
	MaxArgs  int // -1 for unlimited
	Evaluate func(context.Context, EvaluationContext, Args) (value.Value, []point.Point, error)
}

func (e *ExpressionFunction) IsInvalid(args Args) bool {
	len := args.Len()
	if e.MinArgs > len {
		return true
	}

	if e.MaxArgs != -1 && e.MaxArgs < len {
		return true
	}

	return false
}

type Args []Expression

func (a Args) Len() int {
	return len(a)
}

func (a Args) Evaluate(ctx context.Context, deps EvaluationContext, i int) (value.Value, []point.Point, error) {
	if i >= len(a) {
		return value.NewNull(), []point.Point{}, nil
	}

	return a[i].Evaluate(ctx, deps)
}

func (a Args) Identifier(ctx context.Context, deps EvaluationContext, i int) (string, error) {
	v, cfp, err := a.Evaluate(ctx, deps, i)
	if err != nil {
		return "", err
	}

	if len(cfp) > 0 {
		return "", errors.New("unexpected configuration point")
	}

	if v.Kind() != value.Identifier {
		return "", errors.New("expected identifier")
	}

	return v.String(), nil
}
