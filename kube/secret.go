package kube

import (
	"context"
	"fmt"
	"slices"

	"github.com/BSFishy/mora-manager/core"
	"github.com/BSFishy/mora-manager/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Secret struct {
	moduleName string
	identifier string
	value      []byte
}

const secretKey string = "value"

func NewSecret(deps interface {
	core.HasModuleName
}, identifier string, value []byte,
) Resource[corev1.Secret] {
	moduleName := deps.GetModuleName()

	return &Secret{
		moduleName: moduleName,
		identifier: identifier,
		value:      value,
	}
}

func (s *Secret) Name() string {
	return util.SanitizeDNS1123Subdomain(fmt.Sprintf("%s-%s", s.moduleName, s.identifier))
}

func (s *Secret) Get(ctx context.Context, deps KubeContext) (*corev1.Secret, error) {
	return deps.GetClientset().CoreV1().Secrets(namespace(deps)).Get(ctx, s.Name(), metav1.GetOptions{})
}

func (s *Secret) IsValid(ctx context.Context, secret *corev1.Secret) (bool, error) {
	data, ok := secret.Data[secretKey]
	if !ok {
		return false, nil
	}

	if slices.Compare(data, s.value) != 0 {
		return false, nil
	}

	return true, nil
}

func (s *Secret) Delete(ctx context.Context, deps KubeContext) error {
	return deps.GetClientset().CoreV1().Secrets(namespace(deps)).Delete(ctx, s.Name(), metav1.DeleteOptions{})
}

func (s *Secret) Create(ctx context.Context, deps KubeContext) (*corev1.Secret, error) {
	labels := matchLabels(deps, map[string]string{
		"mora.identifier": s.identifier,
	})

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace(deps),
			Name:      s.Name(),
			Labels:    labels,
		},
		Data: map[string][]byte{
			secretKey: s.value,
		},
		// TODO: make this configurable?
		Type: corev1.SecretTypeOpaque,
	}

	return deps.GetClientset().CoreV1().Secrets(namespace(deps)).Create(ctx, secret, metav1.CreateOptions{})
}

func (s *Secret) Ready(secret *corev1.Secret) bool {
	// secrets are immediately available once create returns with no error
	return true
}

func GetSecret(ctx context.Context, deps KubeContext, identifier string) ([]byte, error) {
	secret := NewSecret(deps, identifier, nil)
	res, err := secret.Get(ctx, deps)
	if err != nil {
		return nil, err
	}

	return res.Data[secretKey], nil
}
