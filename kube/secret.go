package kube

import (
	"context"
	"fmt"
	"slices"

	"github.com/BSFishy/mora-manager/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Secret struct {
	moduleName string
	identifier string
	value      []byte
}

const secretKey string = "value"

func NewSecret(ctx context.Context, identifier string, value []byte) Resource[corev1.Secret] {
	moduleName := util.Has(util.GetModuleName(ctx))

	return &Secret{
		moduleName: moduleName,
		identifier: identifier,
		value:      value,
	}
}

func (s *Secret) Name() string {
	return util.SanitizeDNS1123Subdomain(fmt.Sprintf("%s-%s", s.moduleName, s.identifier))
}

func (s *Secret) Get(ctx context.Context, client *kubernetes.Clientset) (*corev1.Secret, error) {
	return client.CoreV1().Secrets(namespace(ctx)).Get(ctx, s.Name(), metav1.GetOptions{})
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

func (s *Secret) Delete(ctx context.Context, client *kubernetes.Clientset) error {
	return client.CoreV1().Secrets(namespace(ctx)).Delete(ctx, s.Name(), metav1.DeleteOptions{})
}

func (s *Secret) Create(ctx context.Context, client *kubernetes.Clientset) (*corev1.Secret, error) {
	labels := matchLabels(ctx, map[string]string{
		"mora.identifier": s.identifier,
	})

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace(ctx),
			Name:      s.Name(),
			Labels:    labels,
		},
		Data: map[string][]byte{
			secretKey: s.value,
		},
		// TODO: make this configurable?
		Type: corev1.SecretTypeOpaque,
	}

	return client.CoreV1().Secrets(namespace(ctx)).Create(ctx, secret, metav1.CreateOptions{})
}

func (s *Secret) Ready(secret *corev1.Secret) bool {
	// secrets are immediately available once create returns with no error
	return true
}
