package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/BSFishy/mora-manager/config"
	"github.com/BSFishy/mora-manager/kube"
	"github.com/BSFishy/mora-manager/model"
	"github.com/BSFishy/mora-manager/state"
	"github.com/BSFishy/mora-manager/util"
)

func (a *App) deploy(d *model.Deployment) {
	ctx := context.Background()
	logger := util.LogFromCtx(ctx)

	logger = logger.With("deployment", d.Id, "environment", d.EnvironmentId)
	ctx = util.WithLogger(ctx, logger)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go a.handleDeployCancel(ctx, cancel, d)

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

		ctx = model.WithUser(ctx, user)

		if err = d.Refresh(ctx, tx); err != nil {
			return fmt.Errorf("refreshing deployment: %w", err)
		}

		if d.Status != model.NotStarted && d.Status != model.Waiting && d.Status != model.InProgress {
			return errors.New("deployment not in valid state")
		}

		if err = d.UpdateStatusDb(ctx, a.db, model.InProgress); err != nil {
			return fmt.Errorf("updating status: %w", err)
		}

		var cfg config.Config
		if err = json.Unmarshal(d.Config, &cfg); err != nil {
			return fmt.Errorf("decoding config: %w", err)
		}

		var state state.State
		if d.State != nil {
			if err = json.Unmarshal(*d.State, &state); err != nil {
				return fmt.Errorf("decoding state: %w", err)
			}
		}

		if err = kube.EnsureNamespace(ctx, a.WithModel(user, environment)); err != nil {
			return fmt.Errorf("ensuring namespace: %w", err)
		}

		services := cfg.Services[state.ServiceIndex:]
		for _, service := range services {
			logger := logger.With("module", service.ModuleName, "service", service.ServiceName)
			ctx := util.WithLogger(ctx, logger)

			runwayCtx := &runwayContext{
				manager:     a.manager,
				clientset:   a.clientset,
				registry:    a.registry,
				user:        user.Username,
				environment: environment.Slug,
				config:      &cfg,
				state:       &state,
				moduleName:  service.ModuleName,
				serviceName: service.ServiceName,
			}

			wm, configPoints, err := service.EvaluateWingman(ctx, runwayCtx)
			if err != nil {
				return fmt.Errorf("evaluating wingman: %w", err)
			}

			if len(configPoints) > 0 {
				if err = d.UpdateStateAndStatus(ctx, tx, model.Waiting, state); err != nil {
					return fmt.Errorf("updating state: %w", err)
				}

				logger.Info("waiting for dynamic wingman config")
				return nil
			}

			if wm != nil {
				mwm := wm.MaterializeWingman(runwayCtx)
				if err = mwm.Deploy(ctx, runwayCtx); err != nil {
					return fmt.Errorf("deploying wingman: %w", err)
				}

				logger.Info("deployed wingman")

				rwm, err := a.manager.FindWingman(ctx, runwayCtx)
				if err != nil {
					return fmt.Errorf("finding wingman: %w", err)
				}

				if rwm != nil {
					cfp, err := rwm.GetConfigPoints(ctx, runwayCtx)
					if err != nil {
						return fmt.Errorf("getting wingman config points: %w", err)
					}

					if len(cfp) > 0 {
						if err = d.UpdateStateAndStatus(ctx, tx, model.Waiting, state); err != nil {
							return fmt.Errorf("updating state: %w", err)
						}

						logger.Info("waiting for dynamic wingman config")
						return nil
					}
				}
			}

			def, configPoints, err := service.Evaluate(ctx, runwayCtx)
			if err != nil {
				return fmt.Errorf("evaluating service: %w", err)
			}

			if len(configPoints) > 0 {
				if err = d.UpdateStateAndStatus(ctx, tx, model.Waiting, state); err != nil {
					return fmt.Errorf("updating state: %w", err)
				}

				logger.Info("waiting for dynamic config")
				return nil
			}

			deployment := def.Materialize(runwayCtx)
			if err = deployment.Deploy(ctx, runwayCtx); err != nil {
				return fmt.Errorf("deploying service: %w", err)
			}

			state.ServiceIndex++
			logger.Info("deployed service")
		}

		if err = d.UpdateStateAndStatus(ctx, tx, model.Success, state); err != nil {
			return fmt.Errorf("updating status to success: %w", err)
		}

		logger.Info("deployment successful")

		return nil
	})
	if err != nil {
		// we make errors crazy with more info. this just checks if the error chain
		// terminates with a context canceled error
		if strings.HasSuffix(err.Error(), context.Canceled.Error()) {
			return
		}

		logger.Error("deployment failed", "err", err)

		if err := d.UpdateStatusDb(ctx, a.db, model.Errored); err != nil {
			logger.Error("updating status to errored", "err", err)
		}
	}
}
