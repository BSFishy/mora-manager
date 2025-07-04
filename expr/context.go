package expr

import (
	"context"

	"github.com/BSFishy/mora-manager/core"
	"github.com/BSFishy/mora-manager/model"
	"github.com/BSFishy/mora-manager/point"
	"github.com/BSFishy/mora-manager/state"
	"github.com/BSFishy/mora-manager/value"
)

type FunctionRegistry interface {
	Evaluate(context.Context, EvaluationContext, string, Args) (value.Value, []point.Point, error)
}

type HasFunctionRegistry interface {
	GetFunctionRegistry() FunctionRegistry
}

type Config interface {
	FindConfig(string, string) *point.Point
}

type HasConfig interface {
	GetConfig() Config
}

type EvaluationContext interface {
	core.HasClientSet
	HasFunctionRegistry
	model.HasEnvironment
	model.HasUser
	state.HasState
	HasConfig
	core.HasModuleName
}
