package config

import (
	"github.com/BSFishy/mora-manager/point"
	"github.com/BSFishy/mora-manager/value"
)

type Config struct {
	Services []ServiceConfig
	Configs  []point.Point
}

func (c *Config) FindConfig(moduleName, identifier string) *point.Point {
	for _, config := range c.Configs {
		if config.ModuleName == moduleName && config.Identifier == identifier {
			return &config
		}
	}

	return nil
}

type MaterializedEnv struct {
	Name  string
	Value value.Value
}
