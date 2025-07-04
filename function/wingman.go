package function

import (
	"context"

	"github.com/BSFishy/mora-manager/core"
	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/model"
	"github.com/BSFishy/mora-manager/point"
	"github.com/BSFishy/mora-manager/state"
	"github.com/BSFishy/mora-manager/value"
)

type HasWingmanManager interface {
	GetWingmanManager() WingmanManager
}

type WingmanManager interface {
	EvaluateFunction(context.Context, interface {
		model.HasUser
		model.HasEnvironment
		core.HasClientSet
		state.HasState
	}, string, expr.Args,
	) (value.Value, []point.Point, error)
}
