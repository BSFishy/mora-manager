package config

import (
	"context"
	"errors"
	"fmt"

	"github.com/BSFishy/mora-manager/api"
	"github.com/BSFishy/mora-manager/core"
	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/point"
	"github.com/BSFishy/mora-manager/state"
	"github.com/BSFishy/mora-manager/util/shlex"
	"github.com/BSFishy/mora-manager/value"
	"k8s.io/client-go/kubernetes"
)

type ServiceWingman struct {
	Image expr.Expression
}

type ServiceConfig struct {
	ModuleName  string
	ServiceName string
	Image       expr.Expression
	Command     *expr.Expression
	Env         []api.Env

	Wingman *ServiceWingman
}

type configFromModulesCtx struct {
	client      kubernetes.Interface
	registry    expr.FunctionRegistry
	user        string
	environment string
	moduleName  string
}

func (c *configFromModulesCtx) GetClientset() kubernetes.Interface {
	return c.client
}

func (c *configFromModulesCtx) GetFunctionRegistry() expr.FunctionRegistry {
	return c.registry
}

func (c *configFromModulesCtx) GetUser() string {
	return c.user
}

func (c *configFromModulesCtx) GetEnvironment() string {
	return c.environment
}

func (c *configFromModulesCtx) GetConfig() expr.Config {
	panic("invalid call to GetConfig")
}

func (c *configFromModulesCtx) GetState() *state.State {
	panic("invalid call to GetState")
}

func (c *configFromModulesCtx) GetModuleName() string {
	return c.moduleName
}

func ServiceConfigFromModules(ctx context.Context, deps interface {
	core.HasClientSet
	expr.HasFunctionRegistry
	core.HasUser
	core.HasEnvironment
}, modules []api.Module,
) ([]ServiceConfig, error) {
	services := make(map[string]ServiceConfig)
	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	for _, module := range modules {
		moduleDeps := &configFromModulesCtx{
			client:      deps.GetClientset(),
			registry:    deps.GetFunctionRegistry(),
			user:        deps.GetUser(),
			environment: deps.GetEnvironment(),
			moduleName:  module.Name,
		}

		for _, service := range module.Services {
			path := fmt.Sprintf("%s/%s", module.Name, service.Name)
			requires, err := service.RequiredServices(ctx, moduleDeps)
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
				Command:     service.Command,
				Env:         service.Env,
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

func (s *ServiceConfig) EvaluateWingman(ctx context.Context, deps expr.EvaluationContext) (*WingmanDefinition, []point.Point, error) {
	if s.Wingman == nil {
		return nil, nil, nil
	}

	configPoints := []point.Point{}

	wingmanImage, wingmanImageCfp, err := s.Wingman.Image.Evaluate(ctx, deps)
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

func (s *ServiceConfig) Evaluate(ctx context.Context, deps expr.EvaluationContext) (*ServiceDefinition, []point.Point, error) {
	configPoints := []point.Point{}

	image, imageCfp, err := s.Image.Evaluate(ctx, deps)
	if err != nil {
		return nil, nil, fmt.Errorf("evaluating image: %w", err)
	}

	if image.Kind() != value.String {
		return nil, nil, errors.New("invalid image property")
	}

	configPoints = append(configPoints, imageCfp...)

	var command []string
	if s.Command != nil {
		cmd, cmdCfp, err := s.Command.Evaluate(ctx, deps)
		if err != nil {
			return nil, nil, fmt.Errorf("evaluating command: %w", err)
		}

		if len(cmdCfp) == 0 && cmd.Kind() != value.String {
			return nil, nil, errors.New("invalid command property")
		}

		configPoints = append(configPoints, cmdCfp...)
		cmdString := cmd.String()

		command, err = shlex.Split(cmdString)
		if err != nil {
			return nil, nil, fmt.Errorf("splitting command: %w", err)
		}
	}

	envs := []MaterializedEnv{}
	for _, e := range s.Env {
		ev, envCfp, err := e.Value.Evaluate(ctx, deps)
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
		Image:   image.String(),
		Command: command,
		Env:     envs,
	}, configPoints, nil
}
