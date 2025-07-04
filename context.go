package main

import (
	"github.com/BSFishy/mora-manager/config"
	"github.com/BSFishy/mora-manager/core"
	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/model"
	"github.com/BSFishy/mora-manager/state"
	"github.com/BSFishy/mora-manager/wingman"
	"k8s.io/client-go/kubernetes"
)

var (
	_ wingman.HasManager       = (*modelContext)(nil)
	_ core.HasClientSet        = (*modelContext)(nil)
	_ expr.HasFunctionRegistry = (*modelContext)(nil)
	_ model.HasUser            = (*modelContext)(nil)
	_ model.HasEnvironment     = (*modelContext)(nil)
)

func (a *App) WithModel(user *model.User, env *model.Environment) *modelContext {
	return &modelContext{
		manager:     a.manager,
		clientset:   a.clientset,
		registry:    a.registry,
		user:        user,
		environment: env,
	}
}

type modelContext struct {
	manager     *wingman.Manager
	clientset   *kubernetes.Clientset
	registry    expr.FunctionRegistry
	user        *model.User
	environment *model.Environment
}

func (m *modelContext) GetWingmanManager() *wingman.Manager {
	return m.manager
}

func (m *modelContext) GetClientset() kubernetes.Interface {
	return m.clientset
}

func (m *modelContext) GetFunctionRegistry() expr.FunctionRegistry {
	return m.registry
}

func (m *modelContext) GetUser() *model.User {
	return m.user
}

func (m *modelContext) GetEnvironment() *model.Environment {
	return m.environment
}

var (
	_ expr.EvaluationContext = (*runwayContext)(nil)
	_ wingman.HasManager     = (*runwayContext)(nil)
	_ model.HasUser          = (*runwayContext)(nil)
	_ model.HasEnvironment   = (*runwayContext)(nil)
	_ core.HasServiceName    = (*runwayContext)(nil)
	_ core.HasClientSet      = (*runwayContext)(nil)
)

type runwayContext struct {
	manager     *wingman.Manager
	clientset   *kubernetes.Clientset
	registry    expr.FunctionRegistry
	user        *model.User
	environment *model.Environment
	config      *config.Config
	state       *state.State
	moduleName  string
	serviceName string
}

func (r *runwayContext) GetWingmanManager() *wingman.Manager {
	return r.manager
}

func (r *runwayContext) GetClientset() kubernetes.Interface {
	return r.clientset
}

func (r *runwayContext) GetFunctionRegistry() expr.FunctionRegistry {
	return r.registry
}

func (r *runwayContext) GetUser() *model.User {
	return r.user
}

func (r *runwayContext) GetEnvironment() *model.Environment {
	return r.environment
}

func (r *runwayContext) GetConfig() expr.Config {
	return r.config
}

func (r *runwayContext) GetState() *state.State {
	return r.state
}

func (r *runwayContext) GetModuleName() string {
	return r.moduleName
}

func (r *runwayContext) GetServiceName() string {
	return r.serviceName
}
