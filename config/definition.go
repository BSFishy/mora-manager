package config

import (
	"fmt"
	"path"

	"github.com/BSFishy/mora-manager/core"
	"github.com/BSFishy/mora-manager/def"
	"github.com/BSFishy/mora-manager/kube"
	"github.com/BSFishy/mora-manager/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

type ConfigMap struct {
	Path string
	File []byte
}

type ServiceDefinition struct {
	Image   string
	Command []string
	Configs []ConfigMap
	Env     []MaterializedEnv
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

	// TODO: make this support also doing full directories and all that
	mounts := make([]corev1.VolumeMount, len(s.Configs))
	volumes := make([]corev1.Volume, len(s.Configs))
	configMaps := make([]kube.Resource[corev1.ConfigMap], len(s.Configs))
	for i, c := range s.Configs {
		config := kube.NewConfigMap(deps, c.Path, c.File)
		mounts[i] = corev1.VolumeMount{
			Name:      config.Name(),
			MountPath: path.Dir(c.Path),
		}

		volumes[i] = corev1.Volume{
			Name: config.Name(),
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: config.Name(),
					},
					Items: []corev1.KeyToPath{
						{
							Key:  kube.ConfigKey,
							Path: path.Base(c.Path),
						},
					},
				},
			},
		}

		configMaps[i] = config
	}

	return &kube.MaterializedService{
		ConfigMaps: configMaps,
		Deployments: []kube.Resource[appsv1.Deployment]{
			kube.NewDeployment(deps, s.Image, s.Command, env, false, "", mounts, volumes),
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
			// TODO: support commands for wingmen
			kube.NewDeployment(deps, w.Image, nil, nil, true, name, nil, nil),
		},
		Services: []kube.Resource[corev1.Service]{
			kube.NewService(deps, true),
		},
	}
}
