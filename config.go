package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/BSFishy/mora-manager/config"
	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/kube"
	"github.com/BSFishy/mora-manager/value"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	Services []ServiceConfig
	Configs  []config.Point
}

func (c *Config) FindConfig(moduleName, identifier string) *config.Point {
	for _, config := range c.Configs {
		if config.ModuleName == moduleName && config.Identifier == identifier {
			return &config
		}
	}

	return nil
}

type ServiceWingman struct {
	Image expr.Expression
}

type ServiceConfig struct {
	ModuleName  string
	ServiceName string
	Image       expr.Expression
	Env         []Env

	Wingman *ServiceWingman
}

func (s *ServiceConfig) FindConfigPoints(ctx context.Context) (config.Points, error) {
	configPoints := []config.Point{}

	if s.Wingman != nil {
		_, cfp, err := s.EvaluateWingman(ctx)
		if err != nil {
			return nil, fmt.Errorf("evaluating wingman: %w", err)
		}

		configPoints = append(configPoints, cfp...)

		wm, err := FindWingman(ctx)
		if err != nil {
			return nil, fmt.Errorf("getting wingman: %w", err)
		}

		if wm != nil {
			cfp, err = wm.GetConfigPoints(ctx)
			if err != nil {
				return nil, fmt.Errorf("getting wingman config points: %w", err)
			}

			configPoints = append(configPoints, cfp...)
		}
	}

	if len(configPoints) == 0 {
		_, cfp, err := s.Evaluate(ctx)
		if err != nil {
			return nil, fmt.Errorf("evaluating service: %w", err)
		}

		configPoints = append(configPoints, cfp...)
	}

	return configPoints, nil
}

func (s *ServiceConfig) EvaluateWingman(ctx context.Context) (*WingmanDefinition, []config.Point, error) {
	if s.Wingman == nil {
		return nil, nil, nil
	}

	configPoints := []config.Point{}

	wingmanImage, wingmanImageCfp, err := s.Wingman.Image.Evaluate(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("evaluating wingman image: %w", err)
	}

	if wingmanImage.Kind() != value.String {
		return nil, nil, errors.New("invalid wingman image property")
	}

	configPoints = append(configPoints, wingmanImageCfp...)

	if len(configPoints) > 0 {
		return nil, configPoints, nil
	}

	return &WingmanDefinition{
		Image: wingmanImage.String(),
	}, nil, nil
}

func (s *ServiceConfig) Evaluate(ctx context.Context) (*ServiceDefinition, []config.Point, error) {
	configPoints := []config.Point{}

	image, imageCfp, err := s.Image.Evaluate(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("evaluating image: %w", err)
	}

	if image.Kind() != value.String {
		return nil, nil, errors.New("invalid image property")
	}

	configPoints = append(configPoints, imageCfp...)

	envs := []MaterializedEnv{}
	for _, e := range s.Env {
		ev, envCfp, err := e.Value.Evaluate(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("evaluating env %s: %w", e.Name, err)
		}

		configPoints = append(configPoints, envCfp...)

		if len(envCfp) == 0 {
			switch ev.Kind() {
			case value.String:
				fallthrough
			case value.Secret:
				envs = append(envs, MaterializedEnv{
					Name:  e.Name,
					Value: ev,
				})
			default:
				return nil, nil, fmt.Errorf("invalid kind for env %s: %s", e.Name, ev.Kind())
			}
		}
	}

	if len(configPoints) > 0 {
		return nil, configPoints, nil
	}

	return &ServiceDefinition{
		Image: image.String(),
		Env:   envs,
	}, configPoints, nil
}

type WingmanDefinition struct {
	Image string
}

func (w *WingmanDefinition) MaterializeWingman(ctx context.Context) *kube.MaterializedService {
	return &kube.MaterializedService{
		Deployments: []kube.Resource[appsv1.Deployment]{
			kube.NewDeployment(ctx, w.Image, nil, true),
		},
		Services: []kube.Resource[corev1.Service]{
			kube.NewService(ctx, true),
		},
	}
}

type MaterializedEnv struct {
	Name  string
	Value value.Value
}

type ServiceDefinition struct {
	Image string
	Env   []MaterializedEnv
}

func (s *ServiceDefinition) Materialize(ctx context.Context) *kube.MaterializedService {
	env := make([]kube.Env, len(s.Env))
	for i, e := range s.Env {
		env[i] = kube.Env{
			Name:  e.Name,
			Value: e.Value,
		}
	}

	return &kube.MaterializedService{
		Deployments: []kube.Resource[appsv1.Deployment]{
			kube.NewDeployment(ctx, s.Image, env, false),
		},
	}
}
