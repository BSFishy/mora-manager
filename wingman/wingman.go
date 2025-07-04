package wingman

import (
	"context"

	"github.com/BSFishy/mora-manager/core"
	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/point"
	"github.com/BSFishy/mora-manager/state"
)

type WingmanContext interface {
	core.HasModuleName
	state.HasState
}

type Wingman interface {
	GetConfigPoints(context.Context, WingmanContext) ([]point.Point, error)
	GetFunctions(context.Context, WingmanContext) (map[string]expr.ExpressionFunction, error)
}
