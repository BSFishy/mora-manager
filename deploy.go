package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/BSFishy/mora-manager/expr"
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
	ctx = expr.WithFunctionRegistry(ctx, a.registry)

	err := a.db.Transact(ctx, func(tx *sql.Tx) error {
		err := d.Lock(ctx, tx)
		if err != nil {
			return err
		}

		environment, err := a.db.GetEnvironment(ctx, d.EnvironmentId)
		if err != nil {
			return fmt.Errorf("getting environment: %w", err)
		}

		ctx = model.WithEnvironment(ctx, environment)

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

		var cfg Config
		if err = json.Unmarshal(d.Config, &cfg); err != nil {
			return fmt.Errorf("decoding config: %w", err)
		}

		ctx = WithConfig(ctx, &cfg)

		var state state.State
		if d.State != nil {
			if err = json.Unmarshal(*d.State, &state); err != nil {
				return fmt.Errorf("decoding state: %w", err)
			}
		}

		ctx = WithState(ctx, &state)

		if err = kube.EnsureNamespace(ctx, a.clientset); err != nil {
			return fmt.Errorf("ensuring namespace: %w", err)
		}

		services := cfg.Services[state.ServiceIndex:]
		for _, service := range services {
			logger := logger.With("module", service.ModuleName, "service", service.ServiceName)
			ctx := util.WithLogger(ctx, logger)

			ctx = util.WithModuleName(ctx, service.ModuleName)
			ctx = util.WithServiceName(ctx, service.ServiceName)

			def, configPoints, err := service.Evaluate(ctx)
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

			if wm := def.MaterializeWingman(ctx); wm != nil {
				if err = wm.Deploy(ctx, a.clientset); err != nil {
					return fmt.Errorf("deploying wingman: %w", err)
				}

				logger.Info("deployed wingman")

				rwm, err := a.FindWingman(ctx)
				if err != nil {
					return fmt.Errorf("finding wingman: %w", err)
				}

				if rwm != nil {
					cfp, err := rwm.GetConfigPoints(ctx)
					if err != nil {
						return fmt.Errorf("getting wingman config points: %w", err)
					}

					// TODO: this feels wrong. why am i adding it to the config? i feel
					// like this should just be something i add directly to the state and
					// be good with it?
					for _, point := range cfp {
						cfg.Configs = append(cfg.Configs, point)
					}

					if len(cfp) > 0 {
						if err = d.UpdateConfig(ctx, tx, cfg); err != nil {
							return fmt.Errorf("updating config: %w", err)
						}

						if err = d.UpdateStateAndStatus(ctx, tx, model.Waiting, state); err != nil {
							return fmt.Errorf("updating state: %w", err)
						}

						logger.Info("waiting for dynamic wingman config")
						return nil
					}
				}
			}

			deployment := def.Materialize(ctx)
			if err = deployment.Deploy(ctx, a.clientset); err != nil {
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
		logger.Error("deployment failed", "err", err)

		if err := d.UpdateStatusDb(ctx, a.db, model.Errored); err != nil {
			logger.Error("updating status to errored", "err", err)
		}
	}
}
