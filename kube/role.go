package kube

import (
	"context"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Role struct {
	name  string
	rules []rbacv1.PolicyRule
}

func NewRole(name string, rules []rbacv1.PolicyRule) Resource[rbacv1.Role] {
	return &Role{
		name:  name,
		rules: rules,
	}
}

func (r *Role) Name() string {
	return r.name
}

func (r *Role) Get(ctx context.Context, deps KubeContext) (*rbacv1.Role, error) {
	return deps.GetClientset().RbacV1().Roles(namespace(deps)).Get(ctx, r.Name(), metav1.GetOptions{})
}

func (r *Role) Manifest(ctx context.Context, deps KubeContext) *rbacv1.Role {
	labels := matchLabels(deps, map[string]string{
		"mora.name": r.name,
	})
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name(),
			Namespace: namespace(deps),
			Labels:    labels,
		},
		Rules: r.rules,
	}

	return role
}

func (r *Role) Delete(ctx context.Context, deps KubeContext) error {
	return deps.GetClientset().RbacV1().Roles(namespace(deps)).Delete(ctx, r.Name(), metav1.DeleteOptions{})
}

func (r *Role) Create(ctx context.Context, deps KubeContext) (*rbacv1.Role, error) {
	role := r.Manifest(ctx, deps)
	return deps.GetClientset().RbacV1().Roles(namespace(deps)).Create(ctx, role, metav1.CreateOptions{})
}

func (r *Role) Ready(role *rbacv1.Role) bool {
	// there may be a small window where the rbac system is still reconciling this
	// resource, but it should be fine generally
	return true
}
