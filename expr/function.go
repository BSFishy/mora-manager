package expr

import (
	"context"
	"errors"
	"sync"

	"github.com/BSFishy/mora-manager/config"
	"github.com/BSFishy/mora-manager/value"
)

type ExpressionFunction struct {
	MinArgs  int
	MaxArgs  int // -1 for unlimited
	Evaluate func(context.Context, Args) (value.Value, []config.Point, error)
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

type FunctionRegistry struct {
	mu      sync.RWMutex
	funcMap map[string]ExpressionFunction
}

func NewFunctionRegistry() *FunctionRegistry {
	return &FunctionRegistry{
		funcMap: map[string]ExpressionFunction{},
	}
}

func (r *FunctionRegistry) Register(name string, fn ExpressionFunction) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.funcMap[name] = fn
}

func (r *FunctionRegistry) Get(name string) (ExpressionFunction, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fn, ok := r.funcMap[name]
	return fn, ok
}

type Args []Expression

func (a Args) Len() int {
	return len(a)
}

func (a Args) Evaluate(ctx context.Context, i int) (value.Value, []config.Point, error) {
	if i >= len(a) {
		return value.NewNull(), []config.Point{}, nil
	}

	return a[i].Evaluate(ctx)
}

func (a Args) Identifier(ctx context.Context, i int) (string, error) {
	v, cfp, err := a.Evaluate(ctx, i)
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
