package kube

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type MaterializedService struct {
	Deployments []Resource[appsv1.Deployment]
	Services    []Resource[corev1.Service]
}

func (m *MaterializedService) Deploy(ctx context.Context, client *kubernetes.Clientset) error {
	for _, deployment := range m.Deployments {
		if err := Deploy(ctx, client, deployment); err != nil {
			return err
		}
	}

	for _, service := range m.Services {
		if err := Deploy(ctx, client, service); err != nil {
			return err
		}
	}

	return nil
}
