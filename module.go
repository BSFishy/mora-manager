package main

import (
	"context"
	"fmt"
	"time"

	"github.com/BSFishy/mora-manager/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Module struct {
	Name     string    `json:"name"`
	Services []Service `json:"services"`
}

type Service struct {
	Name  string     `json:"name"`
	Image Expression `json:"image"`
}

// TODO: this should be a more comprehensive evaluation system where we create a
// dependency graph of services/modules, split everything out into steps then
// action the steps ad hoc as part of the setup process.
func (s *Service) EvaluateImage() string {
	return *s.Image.Atom.String
}

func (s *Service) Deploy(ctx context.Context, clientset *kubernetes.Clientset, moduleName string) error {
	name := util.SanitizeDNS1123Subdomain(fmt.Sprintf("%s-%s", moduleName, s.Name))
	podspec := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  util.SanitizeDNS1123Label(s.Name),
					Image: s.EvaluateImage(),
				},
			},
		},
	}

	_, err := clientset.CoreV1().Pods("default").Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		err = clientset.CoreV1().Pods("default").Delete(ctx, name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("deleting pod: %w", err)
		}

		for {
			_, err := clientset.CoreV1().Pods("default").Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					break
				}

				return fmt.Errorf("checking pod deletion: %w", err)
			}

			time.Sleep(500 * time.Millisecond)
		}
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf("getting pod: %w", err)
	}

	_, err = clientset.CoreV1().Pods("default").Create(ctx, podspec, metav1.CreateOptions{})
	return err
}

type Expression struct {
	Atom *Atom         `json:"atom,omitempty"`
	List *[]Expression `json:"list,omitempty"`
}

type Atom struct {
	Identifier *string `json:"identifier,omitempty"`
	String     *string `json:"string,omitempty"`
	Number     *string `json:"number,omitempty"`
}
