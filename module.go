package main

import (
	"errors"
	"fmt"

	"github.com/BSFishy/mora-manager/util"
)

type Module struct {
	Name     string         `json:"name"`
	Services []Service      `json:"services"`
	Configs  []ModuleConfig `json:"configs"`
}

type ModuleConfig struct {
	// this field doesnt exist on the api but it is useful for passing these
	// around
	ModuleName  string
	Identifier  string
	Name        Expression
	Description *Expression
}

type Service struct {
	Name     string       `json:"name"`
	Image    Expression   `json:"image"`
	Requires []Expression `json:"requires"`
}

type ServiceRef struct {
	Module  string
	Service string
}

func (s *Service) RequiredServices() ([]ServiceRef, error) {
	services := []ServiceRef{}
	for _, service := range s.Requires {
		list := service.List
		if list == nil {
			return nil, errors.New("required service is not a list")
		}

		expr := *list
		if len(expr) != 3 {
			return nil, errors.New("requires function call invalid")
		}

		f := expr[0].asIdentifier()
		if f == nil {
			return nil, errors.New("list function is not an identifier")
		}

		if *f != "service" {
			return nil, errors.New("requires function is not service")
		}

		moduleName := expr[1].asIdentifier()
		if moduleName == nil {
			return nil, errors.New("service reference module name is not an identifier")
		}

		serviceName := expr[2].asIdentifier()
		if serviceName == nil {
			return nil, errors.New("service reference service name is not an identifier")
		}

		services = append(services, ServiceRef{
			Module:  *moduleName,
			Service: *serviceName,
		})
	}

	return services, nil
}

type ListExpression []Expression

func (l ListExpression) TrivialExpression() *Atom {
	if len(l) != 1 {
		return nil
	}

	e := l[0]
	return e.Atom
}

func (l ListExpression) GetFunctionName(ctx FunctionContext) (string, error) {
	if len(l) < 1 {
		return "", errors.New("invalid empty list expression")
	}

	return l[0].EvaluateIdentifier(ctx)
}

func (l ListExpression) Args() Args {
	return Args(l[1:])
}

type Expression struct {
	Atom *Atom           `json:"atom,omitempty"`
	List *ListExpression `json:"list,omitempty"`
}

func (e *Expression) GetConfigPoints(ctx FunctionContext) ([]ConfigPoint, error) {
	util.AssertEnum("invalid expression", e.Atom, e.List)

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

	fn, ok := ctx.Registry.Get(functionName)
	if !ok {
		return nil, errors.New("invalid function")
	}

	args := list.Args()
	if fn.InvalidArgs(args) {
		return nil, errors.New("invalid arguments")
	}

	return fn.GetConfigPoints(ctx, args)
}

func (e *Expression) EvaluateIdentifier(ctx FunctionContext) (string, error) {
	util.AssertEnum("invalid expression", e.Atom, e.List)

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

	fn, ok := ctx.Registry.Get(functionName)
	if !ok {
		return "", errors.New("invalid function")
	}

	args := list.Args()
	if fn.InvalidArgs(args) {
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

func (e *Expression) EvaluateString(ctx FunctionContext) (string, error) {
	util.AssertEnum("invalid expression", e.Atom, e.List)

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

	fn, ok := ctx.Registry.Get(functionName)
	if !ok {
		return "", errors.New("invalid function")
	}

	args := list.Args()
	if fn.InvalidArgs(args) {
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
