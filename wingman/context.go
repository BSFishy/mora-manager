package wingman

import (
	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/state"
)

var _ expr.EvaluationContext = (*wingmanContext)(nil)

type wingmanContext struct {
	moduleName string
	state      *state.State
	registry   *expr.FunctionRegistry
}

func (w *wingmanContext) GetModuleName() string {
	return w.moduleName
}

func (w *wingmanContext) GetState() *state.State {
	return w.state
}

func (w *wingmanContext) GetFunctionRegistry() *expr.FunctionRegistry {
	return w.registry
}

func (w *wingmanContext) GetConfig() expr.Config {
	panic("unimplemented")
}
