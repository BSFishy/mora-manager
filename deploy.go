package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/BSFishy/mora-manager/model"
	"github.com/BSFishy/mora-manager/util"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (a *App) deploy(d *model.Deployment) {
	ctx := context.Background()
	logger := util.LogFromCtx(ctx)

	logger = logger.With("deployment", d.Id)
	ctx = util.WithLogger(ctx, logger)

	// TODO: i might want to have some sort of mechanism to catch on errors and
	// mark the deployment as failed
	util.Protect(ctx, func() error {
		tx, err := d.Lock(ctx, a.db)
		if err != nil {
			return err
		}

		defer d.Unlock(tx)

		environment, err := a.db.GetEnvironment(ctx, d.EnvironmentId)
		if err != nil {
			return fmt.Errorf("getting environment: %w", err)
		}

		user, err := a.db.GetUserById(ctx, environment.UserId)
		if err != nil {
			return fmt.Errorf("getting user: %w", err)
		}

		if err = d.Refresh(ctx, tx); err != nil {
			return fmt.Errorf("refreshing deployment: %w", err)
		}

		if d.Status != model.NotStarted && d.Status != model.InProgress {
			return errors.New("deployment not in valid state")
		}

		if err = d.UpdateStatus(ctx, tx, model.InProgress); err != nil {
			return fmt.Errorf("updating status: %w", err)
		}

		var config Config
		if err = json.Unmarshal(d.Config, &config); err != nil {
			return fmt.Errorf("decoding config: %w", err)
		}

		var state map[string]any
		if d.State != nil {
			if err = json.Unmarshal(*d.State, &state); err != nil {
				return fmt.Errorf("decoding state: %w", err)
			}
		}

		namespace := fmt.Sprintf("%s-%s", user.Username, environment.Slug)
		if err = ensureNamespace(ctx, a.clientset, namespace); err != nil {
			return fmt.Errorf("ensuring namespace: %w", err)
		}

		for _, service := range config.Services {
			configPoints, err := service.FindConfigPoints(config, state)
			if err != nil {
				return fmt.Errorf("finding config points: %w", err)
			}

			if len(configPoints) > 0 {
				if err = d.UpdateStateAndStatus(ctx, tx, model.Waiting, state); err != nil {
					return fmt.Errorf("updating state: %w", err)
				}

				return nil
			}

			def, err := service.Evaluate(state)
			if err != nil {
				return fmt.Errorf("evaluating service: %w", err)
			}

			isValid, err := def.IsValid(ctx, a.clientset, namespace)
			if err != nil {
				return fmt.Errorf("checking if service is valid: %w", err)
			}

			if !isValid {
				if err = def.Deploy(ctx, a.clientset, namespace); err != nil {
					return fmt.Errorf("deploying service: %w", err)
				}
			}
		}

		if err = d.UpdateStateAndStatus(ctx, tx, model.Success, state); err != nil {
			return fmt.Errorf("updating status to success: %w", err)
		}

		return nil
	})
}

func ensureNamespace(ctx context.Context, clientset *kubernetes.Clientset, namespace string) error {
	_, err := clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		return nil
	}

	if k8serror.IsNotFound(err) {
		config := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
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
