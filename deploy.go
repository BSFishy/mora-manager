package main

import (
	"context"
	"database/sql"
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

	logger = logger.With("deployment", d.Id, "environment", d.EnvironmentId)
	ctx = util.WithLogger(ctx, logger)

	err := a.db.Transact(ctx, func(tx *sql.Tx) error {
		err := d.Lock(ctx, tx)
		if err != nil {
			return err
		}

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

		if d.Status != model.NotStarted && d.Status != model.Waiting && d.Status != model.InProgress {
			return errors.New("deployment not in valid state")
		}

		if err = d.UpdateStatus(ctx, tx, model.InProgress); err != nil {
			return fmt.Errorf("updating status: %w", err)
		}

		var config Config
		if err = json.Unmarshal(d.Config, &config); err != nil {
			return fmt.Errorf("decoding config: %w", err)
		}

		var state State
		if d.State != nil {
			if err = json.Unmarshal(*d.State, &state); err != nil {
				return fmt.Errorf("decoding state: %w", err)
			}
		}

		fnCtx := FunctionContext{
			Registry: a.registry,
			Config:   &config,
			State:    &state,
		}

		namespace := fmt.Sprintf("%s-%s", user.Username, environment.Slug)
		if err = ensureNamespace(ctx, a.clientset, namespace); err != nil {
			return fmt.Errorf("ensuring namespace: %w", err)
		}

		services := state.FilterDeployedServices(config.Services)
		for _, service := range services {
			logger := logger.With("module", service.ModuleName, "service", service.ServiceName)
			ctx := util.WithLogger(ctx, logger)

			moduleFnCtx := fnCtx
			moduleFnCtx.ModuleName = service.ModuleName

			configPoints, err := service.FindConfigPoints(moduleFnCtx)
			if err != nil {
				return fmt.Errorf("finding config points: %w", err)
			}

			if len(configPoints) > 0 {
				if err = d.UpdateStateAndStatus(ctx, tx, model.Waiting, state); err != nil {
					return fmt.Errorf("updating state: %w", err)
				}

				logger.Info("waiting for dynamic config")
				return nil
			}

			def, err := service.Evaluate(moduleFnCtx)
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

			state.AddDeployedService(service.ModuleName, service.ServiceName)
			logger.Info("deployed service")
		}

		if err = d.UpdateStateAndStatus(ctx, tx, model.Success, state); err != nil {
			return fmt.Errorf("updating status to success: %w", err)
		}

		return nil
	})
	if err != nil {
		logger.Error("deployment failed", "err", err)

		if err := d.UpdateStatusDb(ctx, a.db, model.Errored); err != nil {
			logger.Error("updating status to errored", "err", err)
		}
	}
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
