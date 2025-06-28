package kube

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/BSFishy/mora-manager/model"
	"github.com/BSFishy/mora-manager/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Resource[T any] interface {
	Name() string
	Get(context.Context, *kubernetes.Clientset) (*T, error)
	IsValid(context.Context, *T) (bool, error)
	Delete(context.Context, *kubernetes.Clientset) error
	Create(context.Context, *kubernetes.Clientset) (*T, error)
	Ready(*T) bool
}

func Deploy[T any](ctx context.Context, client *kubernetes.Clientset, res Resource[T]) error {
	_, err := res.Get(ctx, client)
	if err == nil {
		if err = res.Delete(ctx, client); err != nil {
			return fmt.Errorf("deleting resource: %w", err)
		}

		for {
			_, err = res.Get(ctx, client)
			if errors.IsNotFound(err) {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf("getting resource: %w", err)
	}

	created, err := res.Create(ctx, client)
	if err != nil {
		return fmt.Errorf("creating resource: %w", err)
	}

	for !res.Ready(created) {
		util.LogFromCtx(ctx).Debug("waiting for resource to be ready")
		time.Sleep(500 * time.Millisecond)

		created, err = res.Get(ctx, client)
		if err != nil {
			return fmt.Errorf("re-fetching for readiness: %w", err)
		}
	}

	return nil
}

func namespace(ctx context.Context) string {
	user := util.Has(model.GetUser(ctx))
	env := util.Has(model.GetEnvironment(ctx))

	return fmt.Sprintf("%s-%s", user.Username, env.Slug)
}

func EnsureNamespace(ctx context.Context, clientset *kubernetes.Clientset) error {
	ns := namespace(ctx)
	_, err := clientset.CoreV1().Namespaces().Get(ctx, ns, metav1.GetOptions{})
	if err == nil {
		return nil
	}

	if errors.IsNotFound(err) {
		config := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: ns,
			},
		}

		_, err := clientset.CoreV1().Namespaces().Create(ctx, &config, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("creating namespace: %w", err)
		}

		return nil
	}

	return fmt.Errorf("getting namespace: %w", err)
}

func matchLabels(ctx context.Context, extras map[string]string) map[string]string {
	user := util.Has(model.GetUser(ctx))
	env := util.Has(model.GetEnvironment(ctx))
	moduleName := util.Has(util.GetModuleName(ctx))

	labels := map[string]string{
		"mora.enabled":     "true",
		"mora.user":        user.Username,
		"mora.environment": env.Slug,
		"mora.module":      moduleName,
	}

	serviceName, ok := util.GetServiceName(ctx)
	if ok {
		labels["mora.service"] = serviceName
	}

	maps.Copy(labels, extras)

	return labels
}
