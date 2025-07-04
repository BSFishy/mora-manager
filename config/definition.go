package config

import (
	"github.com/BSFishy/mora-manager/core"
	"github.com/BSFishy/mora-manager/def"
	"github.com/BSFishy/mora-manager/kube"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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
			kube.NewDeployment(deps, s.Image, env, false),
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
	return &kube.MaterializedService{
		Deployments: []kube.Resource[appsv1.Deployment]{
			kube.NewDeployment(deps, w.Image, nil, true),
		},
		Services: []kube.Resource[corev1.Service]{
			kube.NewService(deps, true),
		},
	}
}
