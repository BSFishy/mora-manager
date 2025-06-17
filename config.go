package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/BSFishy/mora-manager/util"
	"github.com/BSFishy/mora-manager/value"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ConfigPoint struct {
	ModuleName  string
	Identifier  string
	Name        string
	Description *string
}

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
	Image Expression
}

type ServiceConfig struct {
	ModuleName  string
	ServiceName string
	Image       Expression

	Wingman *ServiceWingman
}

func (s *ServiceConfig) Evaluate(ctx context.Context) (*ServiceDefinition, []ConfigPoint, error) {
	user := util.Has(GetUser(ctx))
	environment := util.Has(GetEnvironment(ctx))

	configPoints := []ConfigPoint{}

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

func (s *ServiceDefinition) Deployment(namespace string) *KubernetesDeployment {
	return &KubernetesDeployment{
		Namespace:   namespace,
		User:        s.User,
		Environment: s.Environment,
		Module:      s.Module,
		Service:     s.Name,
		Image:       s.Image,
	}
}

func (s *ServiceDefinition) WingmanDeployment(namespace string) *KubernetesDeployment {
	if s.Wingman == nil {
		return nil
	}

	subservice := "wingman"
	return &KubernetesDeployment{
		Namespace:   namespace,
		User:        s.User,
		Environment: s.Environment,
		Module:      s.Module,
		Service:     s.Name,
		Subservice:  &subservice,
		Image:       s.Wingman.Image,
		Ports: []corev1.ServicePort{
			{
				Port:       8080,
				TargetPort: intstr.FromInt(8080),
			},
		},
	}
}
