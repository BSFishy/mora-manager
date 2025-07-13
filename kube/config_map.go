package kube

import (
	"context"
	"fmt"

	"github.com/BSFishy/mora-manager/core"
	"github.com/BSFishy/mora-manager/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigMap struct {
	moduleName  string
	serviceName string
	name        string
	value       []byte
}

const ConfigKey string = "data"

func NewConfigMap(deps interface {
	core.HasModuleName
	core.HasServiceName
}, name string, value []byte,
) Resource[corev1.ConfigMap] {
	moduleName := deps.GetModuleName()
	serviceName := deps.GetServiceName()

	return &ConfigMap{
		moduleName:  moduleName,
		serviceName: serviceName,
		name:        name,
		value:       value,
	}
}

func (c *ConfigMap) Name() string {
	return util.SanitizeDNS1123Subdomain(fmt.Sprintf("%s-%s-%s", c.moduleName, c.serviceName, c.name))
}

func (c *ConfigMap) Get(ctx context.Context, deps KubeContext) (*corev1.ConfigMap, error) {
	return deps.GetClientset().CoreV1().ConfigMaps(namespace(deps)).Get(ctx, c.Name(), metav1.GetOptions{})
}

func (c *ConfigMap) Manifest(ctx context.Context, deps KubeContext) *corev1.ConfigMap {
	labels := matchLabels(deps, map[string]string{
		"name": c.name,
	})

	config := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace(deps),
			Name:      c.Name(),
			Labels:    labels,
		},
		BinaryData: map[string][]byte{
			ConfigKey: c.value,
		},
	}

	return config
}

func (c *ConfigMap) Delete(ctx context.Context, deps KubeContext) error {
	return deps.GetClientset().CoreV1().ConfigMaps(namespace(deps)).Delete(ctx, c.Name(), metav1.DeleteOptions{})
}

func (c *ConfigMap) Create(ctx context.Context, deps KubeContext) (*corev1.ConfigMap, error) {
	config := c.Manifest(ctx, deps)
	return deps.GetClientset().CoreV1().ConfigMaps(namespace(deps)).Create(ctx, config, metav1.CreateOptions{})
}

func (c *ConfigMap) Ready(config *corev1.ConfigMap) bool {
	// config maps are ready immediately
	return true
}
