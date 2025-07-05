package wingman

import (
	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/state"
	"k8s.io/client-go/kubernetes"
)

var _ expr.EvaluationContext = (*wingmanContext)(nil)

type wingmanContext struct {
	client      *kubernetes.Clientset
	registry    expr.FunctionRegistry
	user        string
	environment string
	state       *state.State
	moduleName  string
}

func (w *wingmanContext) GetClientset() kubernetes.Interface {
	return w.client
}

func (w *wingmanContext) GetFunctionRegistry() expr.FunctionRegistry {
	return w.registry
}

func (w *wingmanContext) GetUser() string {
	return w.user
}

func (w *wingmanContext) GetEnvironment() string {
	return w.environment
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
