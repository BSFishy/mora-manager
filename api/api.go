package api

import (
	"context"

	"github.com/BSFishy/mora-manager/core"
	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/point"
	"github.com/BSFishy/mora-manager/state"
	"k8s.io/client-go/kubernetes"
)

type Config struct {
	Modules []Module `json:"modules"`
}

type flattenContext struct {
	client      kubernetes.Interface
	registry    expr.FunctionRegistry
	user        string
	environment string
	moduleName  string
}

func (f *flattenContext) GetClientset() kubernetes.Interface {
	return f.client
}

func (f *flattenContext) GetFunctionRegistry() expr.FunctionRegistry {
	return f.registry
}

func (f *flattenContext) GetUser() string {
	return f.user
}

func (f *flattenContext) GetEnvironment() string {
	return f.environment
}

func (f *flattenContext) GetConfig() expr.Config {
	panic("invalid call to GetConfig")
}

func (f *flattenContext) GetState() *state.State {
	panic("invalid call to GetState")
}

func (f *flattenContext) GetModuleName() string {
	return f.moduleName
}

func (c *Config) FlattenConfigs(ctx context.Context, deps interface {
	core.HasClientSet
	expr.HasFunctionRegistry
	core.HasUser
	core.HasEnvironment
},
) ([]point.Point, error) {
	configs := []point.Point{}
	for _, module := range c.Modules {
		moduleCtx := &flattenContext{
			client:      deps.GetClientset(),
			registry:    deps.GetFunctionRegistry(),
			user:        deps.GetUser(),
			environment: deps.GetEnvironment(),
			moduleName:  module.Name,
		}

		for _, config := range module.Configs {
			point, err := config.ToConfigPoint(ctx, moduleCtx)
			if err != nil {
				return nil, err
			}

			configs = append(configs, *point)
		}
	}

	return configs, nil
}
