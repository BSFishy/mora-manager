package kube

import (
	"context"
	"fmt"
	"slices"

	"github.com/BSFishy/mora-manager/core"
	"github.com/BSFishy/mora-manager/def"
	"github.com/BSFishy/mora-manager/util"
	"github.com/BSFishy/mora-manager/value"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: need to make a new module so that i can pass a ServiceDefinition around
type Deployment struct {
	moduleName  string
	serviceName string
	image       string
	command     []string
	env         []def.Env
	isWingman   bool

	serviceAccount string
}

func NewDeployment(deps interface {
	core.HasModuleName
	core.HasServiceName
}, image string, command []string, env []def.Env, isWingman bool, serviceAccount string,
) Resource[appsv1.Deployment] {
	moduleName := deps.GetModuleName()
	serviceName := deps.GetServiceName()

	return &Deployment{
		moduleName:     moduleName,
		serviceName:    serviceName,
		image:          image,
		command:        command,
		env:            env,
		isWingman:      isWingman,
		serviceAccount: serviceAccount,
	}
}

func (d *Deployment) Name() string {
	name := fmt.Sprintf("%s-%s", d.moduleName, d.serviceName)
	if d.isWingman {
		name = fmt.Sprintf("%s-wingman", name)
	}

	return util.SanitizeDNS1123Subdomain(name)
}

func (d *Deployment) Get(ctx context.Context, deps KubeContext) (*appsv1.Deployment, error) {
	return deps.GetClientset().AppsV1().Deployments(namespace(deps)).Get(ctx, d.Name(), metav1.GetOptions{})
}

func (d *Deployment) IsValid(ctx context.Context, deployment *appsv1.Deployment) (bool, error) {
	if d.serviceAccount != "" && deployment.Spec.Template.Spec.ServiceAccountName != d.serviceAccount {
		return false, nil
	}

	containers := deployment.Spec.Template.Spec.Containers
	if len(containers) != 1 {
		return false, nil
	}

	container := containers[0]
	if container.Image != d.image {
		return false, nil
	}

	if len(d.command) > 0 && !slices.Equal(container.Command, d.command) {
		return false, nil
	}

	if len(container.Env) != len(d.env) {
		return false, nil
	}

	for _, env := range d.env {
		found := false
		for _, ce := range container.Env {
			if ce.Name == env.Name {
				if env.Value.Kind() == value.Secret {
					if ce.ValueFrom == nil || ce.ValueFrom.SecretKeyRef == nil || ce.ValueFrom.SecretKeyRef.LocalObjectReference.Name != env.Value.String() {
						return false, nil
					}
				} else {
					if ce.Value != env.Value.String() {
						return false, nil
					}
				}

				found = true
				break
			}
		}

		if !found {
			return false, nil
		}
	}

	return true, nil
}

func (d *Deployment) Delete(ctx context.Context, deps KubeContext) error {
	return deps.GetClientset().AppsV1().Deployments(namespace(deps)).Delete(ctx, d.Name(), metav1.DeleteOptions{})
}

func (d *Deployment) Create(ctx context.Context, deps KubeContext) (*appsv1.Deployment, error) {
	extras := map[string]string{}
	if d.isWingman {
		extras["mora.wingman"] = "true"
	} else {
		extras["mora.wingman"] = "false"
	}

	env := make([]corev1.EnvVar, len(d.env))
	for i, e := range d.env {
		if e.Value.Kind() == value.Secret {
			env[i] = corev1.EnvVar{
				Name: e.Name,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: e.Value.String(),
						},
						Key: "value",
					},
				},
			}
		} else {
			env[i] = corev1.EnvVar{
				Name:  e.Name,
				Value: e.Value.String(),
			}
		}
	}

	labels := matchLabels(deps, extras)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace(deps),
			Name:      d.Name(),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace(deps),
					Name:      d.Name(),
					Labels:    labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: d.serviceAccount,
					Containers: []corev1.Container{
						{
							Name:    util.SanitizeDNS1123Label(d.Name()),
							Image:   d.image,
							Command: d.command,
							Env:     env,
						},
					},
				},
			},
		},
	}

	return deps.GetClientset().AppsV1().Deployments(namespace(deps)).Create(ctx, deployment, metav1.CreateOptions{})
}

func (d *Deployment) Ready(deployment *appsv1.Deployment) bool {
	// NOTE: this assumes we will never have more than 1 replica. will need to
	// change this when we support that.
	return deployment.Status.ReadyReplicas == 1
}
