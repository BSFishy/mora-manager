package kube

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

type MaterializedService struct {
	Deployments []Resource[appsv1.Deployment]
	Services    []Resource[corev1.Service]
	Secrets     []Resource[corev1.Secret]

	Roles           []Resource[rbacv1.Role]
	RoleBindings    []Resource[rbacv1.RoleBinding]
	ServiceAccounts []Resource[corev1.ServiceAccount]
}

func (m *MaterializedService) Deploy(ctx context.Context, deps KubeContext) error {
	if err := deployAll(ctx, deps, m.Roles); err != nil {
		return err
	}

	if err := deployAll(ctx, deps, m.RoleBindings); err != nil {
		return err
	}

	if err := deployAll(ctx, deps, m.ServiceAccounts); err != nil {
		return err
	}

	if err := deployAll(ctx, deps, m.Secrets); err != nil {
		return err
	}

	if err := deployAll(ctx, deps, m.Deployments); err != nil {
		return err
	}

	if err := deployAll(ctx, deps, m.Services); err != nil {
		return err
	}

	return nil
}

func deployAll[T any](ctx context.Context, deps KubeContext, resources []Resource[T]) error {
	for _, res := range resources {
		if err := Deploy(ctx, deps, res); err != nil {
			return err
		}
	}

	return nil
}
