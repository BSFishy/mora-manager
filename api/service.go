package api

import (
	"context"
	"errors"

	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/state"
	"github.com/BSFishy/mora-manager/value"
)

type Env struct {
	Name  string          `json:"name"`
	Value expr.Expression `json:"value"`
}

type ApiWingman struct {
	Image expr.Expression
}

type Service struct {
	Name     string            `json:"name"`
	Image    expr.Expression   `json:"image"`
	Command  *expr.Expression  `json:"command"`
	Requires []expr.Expression `json:"requires"`
	Wingman  *ApiWingman       `json:"wingman,omitempty"`
	Env      []Env             `json:"env"`
}

func (s *Service) RequiredServices(ctx context.Context, deps expr.EvaluationContext) ([]state.ServiceRef, error) {
	services := []state.ServiceRef{}
	for _, service := range s.Requires {
		v, cfp, err := service.Evaluate(ctx, deps)
		if err != nil {
			return nil, err
		}

		if len(cfp) > 0 {
			return nil, errors.New("unexpected configurable value")
		}

		ref, ok := v.(value.ServiceReferenceValue)
		if !ok {
			return nil, errors.New("invalid requires value")
		}

		// TODO: can i just use value.ServiceReferenceValue here?
		services = append(services, state.ServiceRef{
			Module:  ref.ModuleName,
			Service: ref.ServiceName,
		})
	}

	return services, nil
}
