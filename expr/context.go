package expr

import (
	"github.com/BSFishy/mora-manager/core"
	"github.com/BSFishy/mora-manager/point"
	"github.com/BSFishy/mora-manager/state"
)

type HasFunctionRegistry interface {
	GetFunctionRegistry() *FunctionRegistry
}

type Config interface {
	FindConfig(string, string) *point.Point
}

type HasConfig interface {
	GetConfig() Config
}

type EvaluationContext interface {
	state.HasState
	HasFunctionRegistry
	core.HasModuleName
	HasConfig
}
