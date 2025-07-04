package kube

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/BSFishy/mora-manager/core"
	"github.com/BSFishy/mora-manager/model"
	"github.com/BSFishy/mora-manager/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Resource[T any] interface {
	Name() string
	Get(context.Context, KubeContext) (*T, error)
	IsValid(context.Context, *T) (bool, error)
	Delete(context.Context, KubeContext) error
	Create(context.Context, KubeContext) (*T, error)
	Ready(*T) bool
}

func Deploy[T any](ctx context.Context, deps KubeContext, res Resource[T]) error {
	found, err := res.Get(ctx, deps)
	if err == nil {
		valid, err := res.IsValid(ctx, found)
		if err != nil {
			return fmt.Errorf("checking validity: %w", err)
		}

		if valid {
			return nil
		}

		if err = res.Delete(ctx, deps); err != nil {
			return fmt.Errorf("deleting resource: %w", err)
		}

		for {
			_, err = res.Get(ctx, deps)
			if errors.IsNotFound(err) {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf("getting resource: %w", err)
	}

	created, err := res.Create(ctx, deps)
	if err != nil {
		return fmt.Errorf("creating resource: %w", err)
	}

	for !res.Ready(created) {
		util.LogFromCtx(ctx).Debug("waiting for resource to be ready")
		time.Sleep(500 * time.Millisecond)

		created, err = res.Get(ctx, deps)
		if err != nil {
			return fmt.Errorf("re-fetching for readiness: %w", err)
		}
	}

	return nil
}

func namespace(deps interface {
	model.HasUser
	model.HasEnvironment
},
) string {
	user := deps.GetUser()
	env := deps.GetEnvironment()

	return fmt.Sprintf("%s-%s", user.Username, env.Slug)
}

func EnsureNamespace(ctx context.Context, deps interface {
	model.HasUser
	model.HasEnvironment
	core.HasClientSet
},
) error {
	clientset := deps.GetClientset()
	ns := namespace(deps)

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

func matchLabels(deps interface {
	model.HasUser
	model.HasEnvironment
	core.HasModuleName
}, extras map[string]string,
) map[string]string {
	user := deps.GetUser()
	env := deps.GetEnvironment()
	moduleName := deps.GetModuleName()

	labels := map[string]string{
		"mora.enabled":     "true",
		"mora.user":        user.Username,
		"mora.environment": env.Slug,
		"mora.module":      moduleName,
	}

	if hasServiceName, ok := deps.(core.HasServiceName); ok {
		labels["mora.service"] = hasServiceName.GetServiceName()
	}

	maps.Copy(labels, extras)

	return labels
}
