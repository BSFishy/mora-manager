package kube

import (
	"context"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RoleBinding struct {
	name           string
	role           string
	serviceAccount string
}

func NewRoleBinding(name, role, serviceAccount string) Resource[rbacv1.RoleBinding] {
	return &RoleBinding{
		name:           name,
		role:           role,
		serviceAccount: serviceAccount,
	}
}

func (r *RoleBinding) Name() string {
	return r.name
}

func (r *RoleBinding) Get(ctx context.Context, deps KubeContext) (*rbacv1.RoleBinding, error) {
	return deps.GetClientset().RbacV1().RoleBindings(namespace(deps)).Get(ctx, r.Name(), metav1.GetOptions{})
}

func (r *RoleBinding) Manifest(ctx context.Context, deps KubeContext) *rbacv1.RoleBinding {
	labels := matchLabels(deps, map[string]string{
		"mora.name": r.name,
	})
	binding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name(),
			Namespace: namespace(deps),
			Labels:    labels,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      r.serviceAccount,
				Namespace: namespace(deps),
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     r.role,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	return binding
}

func (r *RoleBinding) Delete(ctx context.Context, deps KubeContext) error {
	return deps.GetClientset().RbacV1().RoleBindings(namespace(deps)).Delete(ctx, r.Name(), metav1.DeleteOptions{})
}

func (r *RoleBinding) Create(ctx context.Context, deps KubeContext) (*rbacv1.RoleBinding, error) {
	binding := r.Manifest(ctx, deps)
	return deps.GetClientset().RbacV1().RoleBindings(namespace(deps)).Create(ctx, binding, metav1.CreateOptions{})
}

func (r *RoleBinding) Ready(binding *rbacv1.RoleBinding) bool {
	// there may be a small window where the rbac system is still reconciling this
	// resource, but it should be fine generally
	return true
}
