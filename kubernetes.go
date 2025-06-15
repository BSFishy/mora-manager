package main

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/BSFishy/mora-manager/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func loadConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		u, err := user.Current()

		var home string
		if err == nil {
			home = u.HomeDir
		} else {
			home = os.Getenv("HOME")
		}

		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

func NewClientset() (*kubernetes.Clientset, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("loading kube config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("creating kubernetes clientset: %w", err)
	}

	return clientset, nil
}

type KubernetesDeployment struct {
	Namespace   string
	User        string
	Environment string
	Module      string
	Service     string
	Subservice  *string
	Image       string
	Ports       []corev1.ServicePort
}

func (k *KubernetesDeployment) Name() string {
	name := fmt.Sprintf("%s-%s", k.Module, k.Service)
	if k.Subservice != nil {
		name = fmt.Sprintf("%s-%s", name, *k.Subservice)
	}

	return util.SanitizeDNS1123Subdomain(name)
}

func (k *KubernetesDeployment) MatchLabels() map[string]string {
	matchLabels := map[string]string{
		"mora.enabled":     "true",
		"mora.user":        k.User,
		"mora.environment": k.Environment,
		"mora.module":      k.Module,
		"mora.service":     k.Service,
	}

	if k.Subservice != nil {
		matchLabels["mora.subservice"] = *k.Subservice
	}

	return matchLabels
}

func (k *KubernetesDeployment) Deploy(ctx context.Context, clientset *kubernetes.Clientset) error {
	isValid, err := k.IsDeploymentValid(ctx, clientset)
	if err != nil {
		return fmt.Errorf("checking if deployment is valid: %w", err)
	}

	if !isValid {
		if err = k.DeployDeployment(ctx, clientset); err != nil {
			return fmt.Errorf("deploying: %w", err)
		}
	}

	isValid, err = k.IsServiceValid(ctx, clientset)
	if err != nil {
		return fmt.Errorf("checking if service is valid: %w", err)
	}

	if !isValid {
		if err = k.DeployService(ctx, clientset); err != nil {
			return fmt.Errorf("deploying service: %w", err)
		}
	}

	return nil
}

func (k *KubernetesDeployment) IsDeploymentValid(ctx context.Context, clientset *kubernetes.Clientset) (bool, error) {
	name := k.Name()

	pod, err := clientset.AppsV1().Deployments(k.Namespace).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		containers := pod.Spec.Template.Spec.Containers
		if len(containers) != 1 {
			return false, nil
		}

		container := containers[0]
		if container.Image != k.Image {
			return false, nil
		}

		return true, nil
	}

	if k8serror.IsNotFound(err) {
		return false, nil
	}

	return false, err
}

func (k *KubernetesDeployment) DeployDeployment(ctx context.Context, clientset *kubernetes.Clientset) error {
	name := k.Name()

	_, err := clientset.AppsV1().Deployments(k.Namespace).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		err = clientset.AppsV1().Deployments(k.Namespace).Delete(ctx, name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("deleting deployment: %w", err)
		}

		for {
			_, err := clientset.AppsV1().Deployments(k.Namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				if k8serror.IsNotFound(err) {
					break
				}

				return fmt.Errorf("checking deployment deletion: %w", err)
			}

			time.Sleep(500 * time.Millisecond)
		}
	} else if !k8serror.IsNotFound(err) {
		return fmt.Errorf("getting deployment: %w", err)
	}

	matchLabels := k.MatchLabels()
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: k.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: matchLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: k.Namespace,
					Labels:    matchLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  util.SanitizeDNS1123Label(k.Service),
							Image: k.Image,
							Ports: []corev1.ContainerPort{},
						},
					},
				},
			},
		},
	}

	_, err = clientset.AppsV1().Deployments(k.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	return nil
}

func (k *KubernetesDeployment) HasService() bool {
	return len(k.Ports) > 0
}

func (k *KubernetesDeployment) IsServiceValid(ctx context.Context, clientset *kubernetes.Clientset) (bool, error) {
	if !k.HasService() {
		return true, nil
	}

	name := k.Name()

	service, err := clientset.CoreV1().Services(k.Namespace).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		ports := service.Spec.Ports
		seen := []int{}
		for _, port := range ports {
			matched := false
			for i, goalPort := range k.Ports {
				if util.Contains(seen, i) {
					continue
				}

				if port.Port == goalPort.Port && port.TargetPort.IntVal == goalPort.TargetPort.IntVal {
					seen = append(seen, i)
					matched = true
					break
				}
			}

			if !matched {
				return false, nil
			}
		}

		return true, nil
	}

	if k8serror.IsNotFound(err) {
		return false, nil
	}

	return false, err
}

func (k *KubernetesDeployment) DeployService(ctx context.Context, clientset *kubernetes.Clientset) error {
	if !k.HasService() {
		return nil
	}

	name := k.Name()

	_, err := clientset.CoreV1().Services(k.Namespace).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		err = clientset.CoreV1().Services(k.Namespace).Delete(ctx, name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("deleting service: %w", err)
		}

		for {
			_, err := clientset.CoreV1().Services(k.Namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				if k8serror.IsNotFound(err) {
					break
				}

				return fmt.Errorf("checking service deletion: %w", err)
			}

			time.Sleep(500 * time.Millisecond)
		}
	} else if !k8serror.IsNotFound(err) {
		return fmt.Errorf("getting service: %w", err)
	}

	matchLabels := k.MatchLabels()
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: k.Namespace,
			Name:      name,
		},
		Spec: corev1.ServiceSpec{
			Selector: matchLabels,
			Ports:    k.Ports,
		},
	}

	_, err = clientset.CoreV1().Services(k.Namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("creating service: %w", err)
	}

	return nil
}
