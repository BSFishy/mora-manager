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

func (l *ListExpression) IsTrivialExpression() bool {
	if len(*l) != 1 {
		return false
	}

	e := (*l)[0]
	return e.Atom != nil
}

func (l *ListExpression) GetFunctionName() (*string, error) {
	if len(*l) < 1 {
		return nil, errors.New("invalid empty list expression")
	}

	e := (*l)[0]
	if e.Atom == nil {
		return nil, errors.New("invalid function call")
	}

	// TODO: do i want this to just return nil if its not an identifier? instead
	// of erroring?
	atom := e.Atom
	return atom.EvaluateIdentifier()
}

type Expression struct {
	Atom *Atom           `json:"atom,omitempty"`
	List *ListExpression `json:"list,omitempty"`
}

func (e *Expression) GetConfigPoints(config Config, state State, moduleName string) ([]ConfigPoint, error) {
	util.AssertEnum("invalid expression", e.Atom, e.List)

	if e.Atom != nil {
		return []ConfigPoint{}, nil
	}

	// expression is a list
	list := e.List
	if list.IsTrivialExpression() {
		return []ConfigPoint{}, nil
	}

	functionName, err := list.GetFunctionName()
	if err != nil {
		return nil, fmt.Errorf("getting function name: %w", err)
	}

	if *functionName == "config" {
		if len(*list) != 2 && len(*list) != 3 {
			return nil, errors.New("invalid config function call")
		}

		configNameExpr := (*list)[1]
		if configNameExpr.Atom == nil {
			return nil, errors.New("invalid config function call")
		}

		configName, err := configNameExpr.Atom.EvaluateIdentifier()
		if err != nil {
			return nil, fmt.Errorf("getting config name: %w", err)
		}

		if len(*list) == 3 {
			moduleNameExpr := (*list)[2]
			if moduleNameExpr.Atom == nil {
				return nil, errors.New("invalid config function call")
			}

			module, err := moduleNameExpr.Atom.EvaluateIdentifier()
			if err != nil {
				return nil, fmt.Errorf("getting module name: %w", err)
			}

			moduleName = *configName
			configName = module
		}

		if cfg := state.FindConfig(moduleName, *configName); cfg != nil {
			return []ConfigPoint{}, nil
		}

		config := config.FindConfig(moduleName, *configName)
		if config == nil {
			return nil, errors.New("invalid config name")
		}

		name, err := config.Name.EvaluateString(state, moduleName)
		if err != nil {
			// realistically, this is a misconfiguration, but ill leave it like this
			return nil, fmt.Errorf("evaluating name: %w", err)
		}

		var description *string
		if config.Description != nil {
			description, err = config.Description.EvaluateString(state, moduleName)
			if err != nil {
				return nil, fmt.Errorf("evaluating description: %w", err)
			}
		}

		return []ConfigPoint{
			{
				ModuleName:  moduleName,
				Identifier:  *configName,
				Name:        *name,
				Description: description,
			},
		}, nil
	}

	return []ConfigPoint{}, nil
}

func (e *Expression) EvaluateString(state State, moduleName string) (*string, error) {
	util.AssertEnum("invalid expression", e.Atom, e.List)

	if e.Atom != nil {
		return e.Atom.EvaluateString()
	}

	list := e.List
	if list.IsTrivialExpression() {
		return (*list)[0].EvaluateString(state, moduleName)
	}

	functionName, err := list.GetFunctionName()
	if err != nil {
		return nil, fmt.Errorf("getting function name: %w", err)
	}

	if *functionName == "config" {
		if len(*list) != 2 && len(*list) != 3 {
			return nil, errors.New("invalid config function call")
		}

		configNameExpr := (*list)[1]
		if configNameExpr.Atom == nil {
			return nil, errors.New("invalid config function call")
		}

		configName, err := configNameExpr.Atom.EvaluateIdentifier()
		if err != nil {
			return nil, fmt.Errorf("getting config name: %w", err)
		}

		if len(*list) == 3 {
			moduleNameExpr := (*list)[2]
			if moduleNameExpr.Atom == nil {
				return nil, errors.New("invalid config function call")
			}

			module, err := moduleNameExpr.Atom.EvaluateIdentifier()
			if err != nil {
				return nil, fmt.Errorf("getting module name: %w", err)
			}

			moduleName = *configName
			configName = module
		}

		if configValue := state.FindConfig(moduleName, *configName); configValue != nil {
			var (
				value string
				ok    bool
			)

			if value, ok = configValue.Value.(string); !ok {
				panic("config string is not a string")
			}

			return &value, nil
		}
	}

	return nil, errors.New("invalid expression")
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

func (a *Atom) EvaluateIdentifier() (*string, error) {
	util.AssertEnum("invalid atom", a.Identifier, a.String, a.Number)

	if a.Identifier == nil {
		return nil, errors.New("atom not identifier")
	}

	return a.Identifier, nil
}

func (a *Atom) EvaluateString() (*string, error) {
	util.AssertEnum("invalid atom", a.Identifier, a.String, a.Number)

	if a.String == nil {
		return nil, errors.New("atom not string")
	}

	return a.String, nil
}
