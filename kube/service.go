package kube

import (
	"context"
	"fmt"

	"github.com/BSFishy/mora-manager/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

type Service struct {
	moduleName  string
	serviceName string
	isWingman   bool
}

func NewService(ctx context.Context, isWingman bool) Resource[corev1.Service] {
	moduleName := util.Has(util.GetModuleName(ctx))
	serviceName := util.Has(util.GetServiceName(ctx))

	return &Service{
		moduleName:  moduleName,
		serviceName: serviceName,
		isWingman:   isWingman,
	}
}

func (s *Service) Name() string {
	name := fmt.Sprintf("%s-%s", s.moduleName, s.serviceName)
	if s.isWingman {
		name = fmt.Sprintf("%s-wingman", name)
	}

	return util.SanitizeDNS1123Subdomain(name)
}

func (s *Service) Get(ctx context.Context, clientset *kubernetes.Clientset) (*corev1.Service, error) {
	return clientset.CoreV1().Services(namespace(ctx)).Get(ctx, s.Name(), metav1.GetOptions{})
}

// TODO: support non-wingman
func (s *Service) IsValid(ctx context.Context, service *corev1.Service) (bool, error) {
	if !s.isWingman {
		return true, nil
	}

	ports := service.Spec.Ports
	if len(ports) != 1 {
		return false, nil
	}

	port := ports[0]
	if port.Port != 8080 || port.TargetPort.IntValue() != 8080 {
		return false, nil
	}

	return true, nil
}

func (s *Service) Delete(ctx context.Context, clientset *kubernetes.Clientset) error {
	return clientset.CoreV1().Services(namespace(ctx)).Delete(ctx, s.Name(), metav1.DeleteOptions{})
}

// TODO: support non-wingman
func (s *Service) Create(ctx context.Context, clientset *kubernetes.Clientset) (*corev1.Service, error) {
	if !s.isWingman {
		return nil, nil
	}

	labels := matchLabels(ctx, map[string]string{
		"mora.wingman": "true",
	})
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace(ctx),
			Name:      s.Name(),
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Port:       8080,
					TargetPort: intstr.FromInt32(8080),
				},
			},
		},
	}

	return clientset.CoreV1().Services(namespace(ctx)).Create(ctx, service, metav1.CreateOptions{})
}

func (s *Service) Ready(service *corev1.Service) bool {
	switch service.Spec.Type {
	case corev1.ServiceTypeLoadBalancer:
		ingress := service.Status.LoadBalancer.Ingress
		for _, i := range ingress {
			if i.IP == "" || i.Hostname == "" {
				return false
			}
		}

		return true
	case corev1.ServiceTypeClusterIP:
		return service.Spec.ClusterIP != "" && service.Spec.ClusterIP != "None"
	case corev1.ServiceTypeNodePort:
		for _, port := range service.Spec.Ports {
			if port.NodePort == 0 {
				return false
			}
		}

		return true
	default:
		return true
	}
}
