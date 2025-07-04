package api

import (
	"context"
	"fmt"

	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/point"
	"github.com/BSFishy/mora-manager/value"
)

type Module struct {
	Name     string         `json:"name"`
	Services []Service      `json:"services"`
	Configs  []ModuleConfig `json:"configs"`
}

type ModuleConfig struct {
	Identifier string
	Name       expr.Expression
	// TODO: maybe this shouldnt be optional?
	Kind        *expr.Expression
	Description *expr.Expression
}

func (m ModuleConfig) ToConfigPoint(ctx context.Context, deps expr.EvaluationContext) (*point.Point, error) {
	nameValue, err := m.Name.ForceEvaluate(ctx, deps)
	if err != nil {
		return nil, fmt.Errorf("evaluating name: %w", err)
	}

	name, err := value.AsString(nameValue)
	if err != nil {
		return nil, err
	}

	var kind string
	if m.Kind != nil {
		kindValue, err := m.Kind.ForceEvaluate(ctx, deps)
		if err != nil {
			return nil, fmt.Errorf("evaluating kind: %w", err)
		}

		// TODO: ideally we check to make sure that this is a valid kind here
		kind, err = value.AsIdentifier(kindValue)
		if err != nil {
			return nil, err
		}
	}

	var description *string
	if m.Description != nil {
		descriptionValue, err := m.Description.ForceEvaluate(ctx, deps)
		if err != nil {
			return nil, fmt.Errorf("evaluating description: %w", err)
		}

		descriptionPtr, err := value.AsString(descriptionValue)
		if err != nil {
			return nil, err
		}

		description = &descriptionPtr
	}

	point := point.Point{
		Identifier:  m.Identifier,
		Name:        name,
		Kind:        point.PointKind(kind),
		Description: description,
	}

	point.Fill(deps)

	return &point, nil
}
