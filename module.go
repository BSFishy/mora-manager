package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/BSFishy/mora-manager/util"
	"github.com/BSFishy/mora-manager/value"
)

type Expression struct {
	Atom *Atom           `json:"atom,omitempty"`
	List *ListExpression `json:"list,omitempty"`
}

func (e *Expression) Evaluate(ctx context.Context) (value.Value, []ConfigPoint, error) {
	util.AssertEnum("invalid expression", e.Atom, e.List)

	if e.Atom != nil {
		v, err := e.Atom.Evaluate()
		return v, []ConfigPoint{}, err
	}

	list := e.List
	if trivial := list.TrivialExpression(); trivial != nil {
		v, err := trivial.Evaluate()
		return v, []ConfigPoint{}, err
	}

	functionName, err := list.GetFunctionName(ctx)
	if err != nil {
		return nil, nil, err
	}

	registry := util.Has(GetFunctionRegistry(ctx))

	fn, ok := registry.Get(functionName)
	if !ok {
		return nil, nil, fmt.Errorf("invalid function: %s", functionName)
	}

	args := list.Args()
	if fn.IsInvalid(args) {
		return nil, nil, fmt.Errorf("invalid arguments for: %s", functionName)
	}

	return fn.Evaluate(ctx, args)
}

type ListExpression []Expression

func (l ListExpression) TrivialExpression() *Atom {
	if len(l) != 1 {
		return nil
	}

	e := l[0]
	return e.Atom
}

func (l ListExpression) GetFunctionName(ctx context.Context) (string, error) {
	if len(l) < 1 {
		return "", errors.New("invalid empty list expression")
	}

	e := l[0]
	v, cfp, err := e.Evaluate(ctx)
	if err != nil {
		return "", fmt.Errorf("evaluating expression: %w", err)
	}

	if len(cfp) > 0 {
		return "", errors.New("unexpected configurable expression")
	}

	if v.Kind() != value.Identifier {
		return "", errors.New("invalid function")
	}

	return v.String(), nil
}

func (l ListExpression) Args() Args {
	return Args(l[1:])
}

type Atom struct {
	Identifier *string `json:"identifier,omitempty"`
	String     *string `json:"string,omitempty"`
	Number     *string `json:"number,omitempty"`
}

func (a *Atom) Evaluate() (value.Value, error) {
	util.AssertEnum("invalid atom", a.Identifier, a.String, a.Number)

	if a.Identifier != nil {
		return value.NewIdentifier(*a.Identifier), nil
	}

	if a.String != nil {
		return value.NewString(*a.String), nil
	}

	i, err := strconv.Atoi(*a.Number)
	if err != nil {
		return nil, err
	}

	return value.NewInteger(i), nil
}
