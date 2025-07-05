package main

import (
	"context"
	"fmt"

	"github.com/BSFishy/mora-manager/config"
	"github.com/BSFishy/mora-manager/core"
	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/point"
	"github.com/BSFishy/mora-manager/wingman"
)

func FindConfigPoints(ctx context.Context, deps interface {
	expr.EvaluationContext
	wingman.HasManager
	core.HasUser
	core.HasEnvironment
	core.HasServiceName
	core.HasClientSet
},
	s *config.ServiceConfig,
) (point.Points, error) {
	configPoints := []point.Point{}

	if s.Wingman != nil {
		_, cfp, err := s.EvaluateWingman(ctx, deps)
		if err != nil {
			return nil, fmt.Errorf("evaluating wingman: %w", err)
		}

		configPoints = append(configPoints, cfp...)

		manager := deps.GetWingmanManager()
		wm, err := manager.FindWingman(ctx, deps)
		if err != nil {
			return nil, fmt.Errorf("getting wingman: %w", err)
		}

		if wm != nil {
			cfp, err = wm.GetConfigPoints(ctx, deps)
			if err != nil {
				return nil, fmt.Errorf("getting wingman config points: %w", err)
			}

			configPoints = append(configPoints, cfp...)
		}
	}

	if len(configPoints) == 0 {
		_, cfp, err := s.Evaluate(ctx, deps)
		if err != nil {
			return nil, fmt.Errorf("evaluating service: %w", err)
		}

		configPoints = append(configPoints, cfp...)
	}

	return configPoints, nil
}
