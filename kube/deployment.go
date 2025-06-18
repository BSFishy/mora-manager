package kube

import (
	"context"
	"fmt"

	"github.com/BSFishy/mora-manager/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Deployment struct {
	moduleName  string
	serviceName string
	image       string
	isWingman   bool
}

func NewDeployment(ctx context.Context, image string, isWingman bool) Resource[appsv1.Deployment] {
	moduleName := util.Has(util.GetModuleName(ctx))
	serviceName := util.Has(util.GetServiceName(ctx))

	return &Deployment{
		moduleName:  moduleName,
		serviceName: serviceName,
		image:       image,
		isWingman:   isWingman,
	}
}

func (d *Deployment) Name() string {
	name := fmt.Sprintf("%s-%s", d.moduleName, d.serviceName)
	if d.isWingman {
		name = fmt.Sprintf("%s-wingman", name)
	}

	return util.SanitizeDNS1123Subdomain(name)
}

func (d *Deployment) Get(ctx context.Context, clientset *kubernetes.Clientset) (*appsv1.Deployment, error) {
	return clientset.AppsV1().Deployments(namespace(ctx)).Get(ctx, d.Name(), metav1.GetOptions{})
}

func (d *Deployment) IsValid(ctx context.Context, deployment *appsv1.Deployment) (bool, error) {
	containers := deployment.Spec.Template.Spec.Containers
	if len(containers) != 1 {
		return false, nil
	}

	container := containers[0]
	if container.Image != d.image {
		return false, nil
	}

	return true, nil
}

func (d *Deployment) Delete(ctx context.Context, clientset *kubernetes.Clientset) error {
	return clientset.AppsV1().Deployments(namespace(ctx)).Delete(ctx, d.Name(), metav1.DeleteOptions{})
}

func (d *Deployment) Create(ctx context.Context, clientset *kubernetes.Clientset) (*appsv1.Deployment, error) {
	extras := map[string]string{}
	if d.isWingman {
		extras["mora.wingman"] = "true"
	}

	labels := matchLabels(ctx, extras)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace(ctx),
			Name:      d.Name(),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace(ctx),
					Name:      d.Name(),
					Labels:    labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  util.SanitizeDNS1123Label(d.Name()),
							Image: d.image,
						},
					},
				},
			},
		},
	}

	return clientset.AppsV1().Deployments(namespace(ctx)).Create(ctx, deployment, metav1.CreateOptions{})
}

func (d *Deployment) Ready(deployment *appsv1.Deployment) bool {
	// NOTE: this assumes we will never have more than 1 replica. will need to
	// change this when we support that.
	return deployment.Status.ReadyReplicas == 1
}
