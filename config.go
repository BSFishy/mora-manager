package main

import (
	"context"
	"fmt"

	"github.com/BSFishy/mora-manager/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ApiConfig struct {
	Modules []Module `json:"modules"`
}

func (c *ApiConfig) FlattenConfigs() []ModuleConfig {
	configs := []ModuleConfig{}
	for _, module := range c.Modules {
		for _, config := range module.Configs {
			config.ModuleName = module.Name
			configs = append(configs, config)
		}
	}

	return configs
}

func (c *ApiConfig) TopologicalSort() ([]ServiceConfig, error) {
	services := make(map[string]ServiceConfig)
	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	for _, module := range c.Modules {
		for _, service := range module.Services {
			path := fmt.Sprintf("%s/%s", module.Name, service.Name)
			requires, err := service.RequiredServices()
			if err != nil {
				return nil, fmt.Errorf("getting required services: %w", err)
			}

			var wingman *ServiceWingman
			if service.Wingman != nil {
				wingman = &ServiceWingman{
					Image: service.Wingman.Image,
				}
			}

			services[path] = ServiceConfig{
				ModuleName:  module.Name,
				ServiceName: service.Name,
				Image:       service.Image,
				Wingman:     wingman,
			}

			for _, dep := range requires {
				depPath := fmt.Sprintf("%s/%s", dep.Module, dep.Service)
				graph[depPath] = append(graph[depPath], path)
				inDegree[path]++
			}

			if _, ok := inDegree[path]; !ok {
				inDegree[path] = 0
			}
		}
	}

	var queue []string
	for path, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, path)
		}
	}

	var result []ServiceConfig
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		result = append(result, services[cur])

		for _, neighbor := range graph[cur] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	return result, nil
}

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

func (s *ServiceConfig) FindConfigPoints(ctx context.Context) ([]ConfigPoint, error) {
	configPoints := []ConfigPoint{}
	image, err := s.Image.GetConfigPoints(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting image config points: %w", err)
	}

	configPoints = append(configPoints, image...)

	if s.Wingman != nil {
		wingmanImage, err := s.Wingman.Image.GetConfigPoints(ctx)
		if err != nil {
			return nil, fmt.Errorf("getting wingman image config points: %w", err)
		}

		configPoints = append(configPoints, wingmanImage...)
	}

	return configPoints, nil
}

func (s *ServiceConfig) Evaluate(ctx context.Context) (*ServiceDefinition, error) {
	user := util.Has(GetUser(ctx))
	environment := util.Has(GetEnvironment(ctx))

	image, err := s.Image.EvaluateString(ctx)
	if err != nil {
		return nil, fmt.Errorf("evaluating image: %w", err)
	}

	var wingman *WingmanDefinition
	if s.Wingman != nil {
		wingmanImage, err := s.Wingman.Image.EvaluateString(ctx)
		if err != nil {
			return nil, fmt.Errorf("evaluating wingman image: %w", err)
		}

		wingman = &WingmanDefinition{
			Image: wingmanImage,
		}
	}

	return &ServiceDefinition{
		User:        user.Username,
		Environment: environment.Slug,
		Module:      s.ModuleName,
		Name:        s.ServiceName,
		Image:       image,

		Wingman: wingman,
	}, nil
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
