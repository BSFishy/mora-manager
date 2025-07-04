package kube

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type MaterializedService struct {
	Deployments []Resource[appsv1.Deployment]
	Services    []Resource[corev1.Service]
	Secrets     []Resource[corev1.Secret]
}

func (m *MaterializedService) Deploy(ctx context.Context, deps KubeContext) error {
	for _, secret := range m.Secrets {
		if err := Deploy(ctx, deps, secret); err != nil {
			return err
		}
	}

	for _, deployment := range m.Deployments {
		if err := Deploy(ctx, deps, deployment); err != nil {
			return err
		}
	}

	for _, service := range m.Services {
		if err := Deploy(ctx, deps, service); err != nil {
			return err
		}
	}

	return nil
}
