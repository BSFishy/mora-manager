package main

import (
	"context"
	"fmt"
	"time"

	"github.com/BSFishy/mora-manager/util"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ApiConfig struct {
	Modules []Module `json:"modules"`
}

func (c *ApiConfig) FlattenConfigs() []ModuleConfig {
	configs := []ModuleConfig{}
	for _, module := range c.Modules {
		for _, config := range module.Configs {
			configs = append(configs, config.WithModuleName(module.Name))
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

			services[path] = ServiceConfig{
				ModuleName:  module.Name,
				ServiceName: service.Name,
				Image:       service.Image,
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
	Name        string
	Description *string
}

type Config struct {
	Services []ServiceConfig
	Configs  []ModuleConfig
}

func (c *Config) FindConfig(name string) *ModuleConfig {
	for _, config := range c.Configs {
		if config.Name == name {
			return &config
		}
	}

	return nil
}

type ServiceConfig struct {
	ModuleName  string
	ServiceName string
	Image       Expression
}

func (s *ServiceConfig) FindConfigPoints(config Config, state State) ([]ConfigPoint, error) {
	configPoints := []ConfigPoint{}
	image, err := s.Image.GetConfigPoints(config, state, s.ModuleName)
	if err != nil {
		return nil, fmt.Errorf("getting image config points: %w", err)
	}

	configPoints = append(configPoints, image...)
	return configPoints, nil
}

func (s *ServiceConfig) Evaluate(state State) (*ServiceDefinition, error) {
	name := fmt.Sprintf("%s_%s", s.ModuleName, s.ServiceName)
	image, err := s.Image.EvaluateString(state, s.ModuleName)
	if err != nil {
		return nil, fmt.Errorf("evaluating image: %w", err)
	}

	return &ServiceDefinition{
		Name:  name,
		Image: *image,
	}, nil
}

type ServiceDefinition struct {
	Name  string
	Image string
}

// TODO: probably check a deployment or something?
func (s *ServiceDefinition) IsValid(ctx context.Context, clientset *kubernetes.Clientset, namespace string) (bool, error) {
	pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, s.Name, metav1.GetOptions{})
	if err == nil {
		containers := pod.Spec.Containers
		if len(containers) != 1 {
			return false, nil
		}

		container := containers[0]
		if container.Image != s.Image {
			return false, nil
		}

		return true, nil
	}

	if k8serror.IsNotFound(err) {
		return false, nil
	}

	return false, err
}

func (s *ServiceDefinition) Deploy(ctx context.Context, clientset *kubernetes.Clientset, namespace string) error {
	name := util.SanitizeDNS1123Subdomain(s.Name)

	_, err := clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		err = clientset.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("deleting pod: %w", err)
		}

		for {
			_, err := clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				if k8serror.IsNotFound(err) {
					break
				}

				return fmt.Errorf("checking pod deletion: %w", err)
			}

			time.Sleep(500 * time.Millisecond)
		}
	} else if !k8serror.IsNotFound(err) {
		return fmt.Errorf("getting pod: %w", err)
	}

	podspec := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  util.SanitizeDNS1123Label(s.Name),
					Image: s.Image,
				},
			},
		},
	}

	_, err = clientset.CoreV1().Pods(namespace).Create(ctx, podspec, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}
