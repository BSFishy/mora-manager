package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/BSFishy/mora-manager/util"
)

type Expression struct {
	Atom *Atom           `json:"atom,omitempty"`
	List *ListExpression `json:"list,omitempty"`
}

func (e *Expression) GetConfigPoints(ctx context.Context) ([]ConfigPoint, error) {
	util.AssertEnum("invalid expression", e.Atom, e.List)

	registry := util.Has(GetFunctionRegistry(ctx))

	if e.Atom != nil {
		return []ConfigPoint{}, nil
	}

	// expression is a list
	list := e.List
	if atom := list.TrivialExpression(); atom != nil {
		return []ConfigPoint{}, nil
	}

	functionName, err := list.GetFunctionName(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting function name: %w", err)
	}

	fn, ok := registry.Get(functionName)
	if !ok {
		return nil, errors.New("invalid function")
	}

	args := list.Args()
	if fn.IsInvalid(args) {
		return nil, errors.New("invalid arguments")
	}

	return fn.GetConfigPoints(ctx, args)
}

func (e *Expression) EvaluateIdentifier(ctx context.Context) (string, error) {
	util.AssertEnum("invalid expression", e.Atom, e.List)

	registry := util.Has(GetFunctionRegistry(ctx))

	if e.Atom != nil {
		return e.Atom.EvaluateIdentifier()
	}

	list := e.List
	if atom := list.TrivialExpression(); atom != nil {
		return atom.EvaluateIdentifier()
	}

	functionName, err := list.GetFunctionName(ctx)
	if err != nil {
		return "", fmt.Errorf("getting function name: %w", err)
	}

	fn, ok := registry.Get(functionName)
	if !ok {
		return "", errors.New("invalid function")
	}

	args := list.Args()
	if fn.IsInvalid(args) {
		return "", errors.New("invalid arguments")
	}

	value, err := fn.Evaluate(ctx, args)
	if err != nil {
		return "", err
	}

	returnValue := value.Identifier()
	if returnValue == nil {
		return "", errors.New("function return is not an identifier")
	}

	return *returnValue, nil
}

func (e *Expression) EvaluateString(ctx context.Context) (string, error) {
	util.AssertEnum("invalid expression", e.Atom, e.List)

	registry := util.Has(GetFunctionRegistry(ctx))

	if e.Atom != nil {
		return e.Atom.EvaluateString()
	}

	list := e.List
	if atom := list.TrivialExpression(); atom != nil {
		return atom.EvaluateString()
	}

	functionName, err := list.GetFunctionName(ctx)
	if err != nil {
		return "", fmt.Errorf("getting function name: %w", err)
	}

	fn, ok := registry.Get(functionName)
	if !ok {
		return "", errors.New("invalid function")
	}

	args := list.Args()
	if fn.IsInvalid(args) {
		return "", errors.New("invalid arguments")
	}

	value, err := fn.Evaluate(ctx, args)
	if err != nil {
		return "", err
	}

	returnValue := value.String()
	if returnValue == nil {
		return "", errors.New("function return is not a string")
	}

	return *returnValue, nil
}

func (e *Expression) asIdentifier() *string {
	if e.Atom == nil {
		return nil
	}

	return e.Atom.Identifier
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

	return l[0].EvaluateIdentifier(ctx)
}

func (l ListExpression) Args() Args {
	return Args(l[1:])
}

type Atom struct {
	Identifier *string `json:"identifier,omitempty"`
	String     *string `json:"string,omitempty"`
	Number     *string `json:"number,omitempty"`
}

func (a *Atom) EvaluateIdentifier() (string, error) {
	util.AssertEnum("invalid atom", a.Identifier, a.String, a.Number)

	if a.Identifier == nil {
		return "", errors.New("atom not identifier")
	}

	return *a.Identifier, nil
}

func (a *Atom) EvaluateString() (string, error) {
	util.AssertEnum("invalid atom", a.Identifier, a.String, a.Number)

	if a.String == nil {
		return "", errors.New("atom not string")
	}

	return *a.String, nil
}
