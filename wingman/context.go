package wingman

import (
	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/model"
	"github.com/BSFishy/mora-manager/state"
	"k8s.io/client-go/kubernetes"
)

var _ expr.EvaluationContext = (*wingmanContext)(nil)

type wingmanContext struct {
	registry   expr.FunctionRegistry
	state      *state.State
	moduleName string
}

func (w *wingmanContext) GetClientset() kubernetes.Interface {
	panic("unimplemented")
}

func (w *wingmanContext) GetFunctionRegistry() expr.FunctionRegistry {
	return w.registry
}

func (w *wingmanContext) GetUser() *model.User {
	panic("unimplemented")
}

func (w *wingmanContext) GetEnvironment() *model.Environment {
	panic("unimplemented")
}

func (w *wingmanContext) GetConfig() expr.Config {
	panic("unimplemented")
}

func (w *wingmanContext) GetState() *state.State {
	return w.state
}

func (w *wingmanContext) GetModuleName() string {
	return w.moduleName
}
