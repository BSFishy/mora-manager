package api

import (
	"context"

	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/point"
	"github.com/BSFishy/mora-manager/state"
)

type Config struct {
	Modules []Module `json:"modules"`
}

type flattenContext struct {
	registry   *expr.FunctionRegistry
	moduleName string
}

func (f *flattenContext) GetFunctionRegistry() *expr.FunctionRegistry {
	return f.registry
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

func (c *Config) FlattenConfigs(ctx context.Context, deps expr.HasFunctionRegistry) ([]point.Point, error) {
	configs := []point.Point{}
	for _, module := range c.Modules {
		moduleCtx := &flattenContext{
			registry:   deps.GetFunctionRegistry(),
			moduleName: module.Name,
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
