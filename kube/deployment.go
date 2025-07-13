package kube

import (
	"context"
	"fmt"

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
	mounts      []corev1.VolumeMount
	volumes     []corev1.Volume
	isWingman   bool

	serviceAccount string
}

func NewDeployment(deps interface {
	core.HasModuleName
	core.HasServiceName
}, image string, command []string, env []def.Env, isWingman bool, serviceAccount string, mounts []corev1.VolumeMount, volumes []corev1.Volume,
) Resource[appsv1.Deployment] {
	moduleName := deps.GetModuleName()
	serviceName := deps.GetServiceName()

	return &Deployment{
		moduleName:     moduleName,
		serviceName:    serviceName,
		image:          image,
		command:        command,
		env:            env,
		mounts:         mounts,
		volumes:        volumes,
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

func (d *Deployment) Manifest(ctx context.Context, deps KubeContext) *appsv1.Deployment {
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
			Labels:    labels,
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
							Name:         util.SanitizeDNS1123Label(d.Name()),
							Image:        d.image,
							Command:      d.command,
							Env:          env,
							VolumeMounts: d.mounts,
						},
					},
					Volumes: d.volumes,
				},
			},
		},
	}

	return deployment
}

func (d *Deployment) Delete(ctx context.Context, deps KubeContext) error {
	return deps.GetClientset().AppsV1().Deployments(namespace(deps)).Delete(ctx, d.Name(), metav1.DeleteOptions{})
}

func (d *Deployment) Create(ctx context.Context, deps KubeContext) (*appsv1.Deployment, error) {
	deployment := d.Manifest(ctx, deps)
	return deps.GetClientset().AppsV1().Deployments(namespace(deps)).Create(ctx, deployment, metav1.CreateOptions{})
}

func (d *Deployment) Ready(deployment *appsv1.Deployment) bool {
	replicas := int32(1)
	if deployment.Spec.Replicas != nil {
		replicas = *deployment.Spec.Replicas
	}

	return deployment.Status.ReadyReplicas == replicas
}
