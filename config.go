package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/kube"
	"github.com/BSFishy/mora-manager/model"
	"github.com/BSFishy/mora-manager/util"
	"github.com/BSFishy/mora-manager/value"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	Services []ServiceConfig
	Configs  []ModuleConfig
}

func (c *Config) FindConfig(moduleName, identifier string) *ModuleConfig {
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

	Wingman *ServiceWingman
}

func (s *ServiceConfig) Evaluate(ctx context.Context) (*ServiceDefinition, []value.ConfigPoint, error) {
	user := util.Has(model.GetUser(ctx))
	environment := util.Has(model.GetEnvironment(ctx))

	configPoints := []value.ConfigPoint{}

	image, imageCfp, err := s.Image.Evaluate(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("evaluating image: %w", err)
	}

	if image.Kind() != value.String {
		return nil, nil, errors.New("invalid image property")
	}

	configPoints = append(configPoints, imageCfp...)

	var wingman *WingmanDefinition
	if s.Wingman != nil {
		wingmanImage, wingmanImageCfp, err := s.Wingman.Image.Evaluate(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("evaluating wingman image: %w", err)
		}

		if wingmanImage.Kind() != value.String {
			return nil, nil, errors.New("invalid wingman image property")
		}

		configPoints = append(configPoints, wingmanImageCfp...)

		wingman = &WingmanDefinition{
			Image: wingmanImage.String(),
		}
	}

	if len(configPoints) > 0 {
		return nil, configPoints, nil
	}

	return &ServiceDefinition{
		User:        user.Username,
		Environment: environment.Slug,
		Module:      s.ModuleName,
		Name:        s.ServiceName,
		Image:       image.String(),

		Wingman: wingman,
	}, configPoints, nil
}

type WingmanDefinition struct {
	Image string
}

type ServiceDefinition struct {
	User        string
	Environment string
	Module      string
	Name        string
	Image       string

	Wingman *WingmanDefinition
}

func (s *ServiceDefinition) Materialize(ctx context.Context) *kube.MaterializedService {
	return &kube.MaterializedService{
		Deployments: []kube.Resource[appsv1.Deployment]{
			kube.NewDeployment(ctx, s.Image, false),
		},
	}
}

func (s *ServiceDefinition) MaterializeWingman(ctx context.Context) *kube.MaterializedService {
	if s.Wingman == nil {
		return &kube.MaterializedService{}
	}

	return &kube.MaterializedService{
		Deployments: []kube.Resource[appsv1.Deployment]{
			kube.NewDeployment(ctx, s.Wingman.Image, true),
		},
		Services: []kube.Resource[corev1.Service]{
			kube.NewService(ctx, true),
		},
	}
}
