package config

import (
	"fmt"

	"github.com/BSFishy/mora-manager/core"
	"github.com/BSFishy/mora-manager/def"
	"github.com/BSFishy/mora-manager/kube"
	"github.com/BSFishy/mora-manager/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

type ServiceDefinition struct {
	Image string
	Env   []MaterializedEnv
}

func (s *ServiceDefinition) Materialize(deps interface {
	core.HasModuleName
	core.HasServiceName
},
) *kube.MaterializedService {
	env := make([]def.Env, len(s.Env))
	for i, e := range s.Env {
		env[i] = def.Env{
			Name:  e.Name,
			Value: e.Value,
		}
	}

	return &kube.MaterializedService{
		Deployments: []kube.Resource[appsv1.Deployment]{
			kube.NewDeployment(deps, s.Image, env, false, ""),
		},
	}
}

type WingmanDefinition struct {
	Image string
}

func (w *WingmanDefinition) MaterializeWingman(deps interface {
	core.HasModuleName
	core.HasServiceName
},
) *kube.MaterializedService {
	moduleName := deps.GetModuleName()
	serviceName := deps.GetServiceName()

	// we're gonna reuse this name across most of the resources. no particular
	// reason, just no real reason to use a bunch of different names.
	name := util.SanitizeDNS1123Subdomain(fmt.Sprintf("%s-%s-wingman", moduleName, serviceName))

	return &kube.MaterializedService{
		Roles: []kube.Resource[rbacv1.Role]{
			kube.NewRole(name, []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"secrets"},
					Verbs:     []string{"get", "list", "create", "update", "delete"},
				},
			}),
		},
		RoleBindings: []kube.Resource[rbacv1.RoleBinding]{
			kube.NewRoleBinding(name, name, name),
		},
		ServiceAccounts: []kube.Resource[corev1.ServiceAccount]{
			kube.NewServiceAccount(name),
		},
		Deployments: []kube.Resource[appsv1.Deployment]{
			kube.NewDeployment(deps, w.Image, nil, true, name),
		},
		Services: []kube.Resource[corev1.Service]{
			kube.NewService(deps, true),
		},
	}
}
