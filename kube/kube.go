package kube

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/BSFishy/mora-manager/core"
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

func pollReady[T any](ctx context.Context, deps KubeContext, res Resource[T], value *T) error {
	for !res.Ready(value) {
		util.LogFromCtx(ctx).Debug("waiting for resource to be ready")
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}

		var err error
		value, err = res.Get(ctx, deps)
		if err != nil {
			return fmt.Errorf("re-fetching for readiness: %w", err)
		}
	}

	return nil
}

func Deploy[T any](ctx context.Context, deps KubeContext, res Resource[T]) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	found, err := res.Get(ctx, deps)
	if err == nil {
		valid, err := res.IsValid(ctx, found)
		if err != nil {
			return fmt.Errorf("checking validity: %w", err)
		}

		if valid {
			return pollReady(ctx, deps, res, found)
		}

		if err = res.Delete(ctx, deps); err != nil {
			return fmt.Errorf("deleting resource: %w", err)
		}

		for {
			_, err = res.Get(ctx, deps)
			if errors.IsNotFound(err) {
				break
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(500 * time.Millisecond):
			}
		}
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf("getting resource: %w", err)
	}

	created, err := res.Create(ctx, deps)
	if err != nil {
		return fmt.Errorf("creating resource: %w", err)
	}

	return pollReady(ctx, deps, res, created)
}

func namespace(deps interface {
	core.HasUser
	core.HasEnvironment
},
) string {
	user := deps.GetUser()
	env := deps.GetEnvironment()

	return fmt.Sprintf("%s-%s", user, env)
}

func EnsureNamespace(ctx context.Context, deps interface {
	core.HasUser
	core.HasEnvironment
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
	core.HasUser
	core.HasEnvironment
	core.HasModuleName
}, extras map[string]string,
) map[string]string {
	user := deps.GetUser()
	env := deps.GetEnvironment()
	moduleName := deps.GetModuleName()

	labels := map[string]string{
		"mora.enabled":     "true",
		"mora.user":        user,
		"mora.environment": env,
		"mora.module":      moduleName,
	}

	if hasServiceName, ok := deps.(core.HasServiceName); ok {
		labels["mora.service"] = hasServiceName.GetServiceName()
	}

	maps.Copy(labels, extras)

	return labels
}
