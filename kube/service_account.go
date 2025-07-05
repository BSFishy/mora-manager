package kube

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceAccount struct {
	name string
}

func NewServiceAccount(name string) Resource[corev1.ServiceAccount] {
	return &ServiceAccount{
		name: name,
	}
}

func (s *ServiceAccount) Name() string {
	return s.name
}

func (s *ServiceAccount) Get(ctx context.Context, deps KubeContext) (*corev1.ServiceAccount, error) {
	return deps.GetClientset().CoreV1().ServiceAccounts(namespace(deps)).Get(ctx, s.Name(), metav1.GetOptions{})
}

func (s *ServiceAccount) IsValid(ctx context.Context, account *corev1.ServiceAccount) (bool, error) {
	return true, nil
}

func (s *ServiceAccount) Delete(ctx context.Context, deps KubeContext) error {
	return deps.GetClientset().CoreV1().ServiceAccounts(namespace(deps)).Delete(ctx, s.Name(), metav1.DeleteOptions{})
}

func (s *ServiceAccount) Create(ctx context.Context, deps KubeContext) (*corev1.ServiceAccount, error) {
	labels := matchLabels(deps, map[string]string{
		"mora.name": s.name,
	})
	account := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name(),
			Namespace: namespace(deps),
			Labels:    labels,
		},
	}

	return deps.GetClientset().CoreV1().ServiceAccounts(namespace(deps)).Create(ctx, account, metav1.CreateOptions{})
}

func (s *ServiceAccount) Ready(account *corev1.ServiceAccount) bool {
	// there may be a small window where the rbac system is still reconciling this
	// resource, but it should be fine generally
	return true
}
