package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/BSFishy/mora-manager/model"
	"github.com/BSFishy/mora-manager/router"
	statepkg "github.com/BSFishy/mora-manager/state"
	"github.com/BSFishy/mora-manager/templates"
	"github.com/BSFishy/mora-manager/util"
)

type DeploymentResponse struct {
	Id string `json:"id"`
}

func (a *App) createDeployment(w http.ResponseWriter, req *http.Request) error {
	var config ApiConfig
	if err := json.NewDecoder(req.Body).Decode(&config); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	ctx := req.Context()
	user, _ := GetUser(ctx)

	params := router.Params(req)
	environmentSlug := params["slug"]

	environment, err := a.db.GetEnvironmentBySlug(ctx, user.Id, environmentSlug)
	if err != nil {
		return fmt.Errorf("getting environment: %w", err)
	}

	if environment == nil || environment.UserId != user.Id {
		http.NotFound(w, req)
		return nil
	}

	if err = environment.CancelInProgressDeployments(ctx, a.db); err != nil {
		return fmt.Errorf("cancelling deployments: %w", err)
	}

	services, err := config.TopologicalSort()
	if err != nil {
		return fmt.Errorf("sorting services: %w", err)
	}

	configs := config.FlattenConfigs()

	deployment, err := environment.NewDeployment(ctx, a.db, Config{
		Services: services,
		Configs:  configs,
	})
	if err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	go a.deploy(deployment)

	return json.NewEncoder(w).Encode(DeploymentResponse{
		Id: deployment.Id,
	})
}

func (a *App) deploymentHtmxRoute(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user, _ := GetUser(ctx)

	environments, err := a.db.GetUserEnvironments(ctx, user.Id)
	if err != nil {
		return fmt.Errorf("getting environments: %w", err)
	}

	deployments, err := a.db.GetDeployments(ctx, environments)
	if err != nil {
		return fmt.Errorf("getting deployments: %w", err)
	}

	return templates.DashboardDeployments(environments, deployments).Render(ctx, w)
}

func (a *App) updateDeploymentConfigHtmxRoute(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	logger := util.LogFromCtx(ctx)

	user, _ := GetUser(ctx)

	params := router.Params(r)
	id := params["id"]

	logger = logger.With("deployment", id)
	ctx = util.WithLogger(ctx, logger)

	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parsing form: %w", err)
	}

	d, err := a.db.GetDeployment(ctx, id)
	if err != nil {
		http.NotFound(w, r)
		return nil
	}

	logger = logger.With("environment", d.EnvironmentId)
	ctx = util.WithLogger(ctx, logger)

	env, err := a.db.GetEnvironment(ctx, d.EnvironmentId)
	if err != nil {
		http.NotFound(w, r)
		logger.Error("failed to get environment in update deployment config")
		return nil
	}

	if env.UserId != user.Id {
		http.NotFound(w, r)
		return nil
	}

	err = a.db.Transact(ctx, func(tx *sql.Tx) error {
		err := d.Lock(ctx, tx)
		if err != nil {
			return fmt.Errorf("taking deployment lock: %w", err)
		}

		moduleNames := r.Form["module_name"]
		identifiers := r.Form["identifier"]
		values := r.Form["value"]

		if len(moduleNames) != len(identifiers) || len(moduleNames) != len(values) {
			w.WriteHeader(http.StatusBadRequest)
			return nil
		}

		var config Config
		if err := json.Unmarshal(d.Config, &config); err != nil {
			return fmt.Errorf("decoding config: %w", err)
		}

		var state statepkg.State
		if d.State != nil {
			if err := json.Unmarshal(*d.State, &state); err != nil {
				return fmt.Errorf("decoding state: %w", err)
			}
		}

		for i := range moduleNames {
			moduleName := moduleNames[i]
			identifier := identifiers[i]
			value := values[i]

			if cfg := config.FindConfig(moduleName, identifier); cfg == nil {
				w.WriteHeader(http.StatusBadRequest)
				return nil
			}

			state.Configs = append(state.Configs, statepkg.StateConfig{
				ModuleName: moduleName,
				Name:       identifier,
				Value:      value,
			})
		}

		if err := d.UpdateState(ctx, tx, state); err != nil {
			return fmt.Errorf("updating state: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	go a.deploy(d)

	return nil
}

func (a *App) deploymentPage(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user, _ := GetUser(ctx)

	params := router.Params(r)
	id := params["id"]

	deployment, err := a.db.GetDeployment(ctx, id)
	if err != nil {
		return fmt.Errorf("getting deployment: %w", err)
	}

	if deployment == nil {
		http.NotFound(w, r)
		return nil
	}

	environment, err := a.db.GetEnvironment(ctx, deployment.EnvironmentId)
	if err != nil {
		return fmt.Errorf("getting environment: %w", err)
	}

	if environment.UserId != user.Id {
		http.NotFound(w, r)
		return nil
	}

	var configPoints []templates.ConfigPoint
	if deployment.Status == model.Waiting {
		var config Config
		if err = json.Unmarshal(deployment.Config, &config); err != nil {
			return fmt.Errorf("decoding config: %w", err)
		}

		var state statepkg.State
		if deployment.State != nil {
			if err = json.Unmarshal(*deployment.State, &state); err != nil {
				return fmt.Errorf("decoding state: %w", err)
			}
		}

		fnCtx := FunctionContext{
			Registry: a.registry,
			Config:   &config,
			State:    &state,
		}

		services := config.Services[state.ServiceIndex:]
		if len(services) > 0 {
			service := services[0]
			moduleFnCtx := fnCtx
			moduleFnCtx.ModuleName = service.ModuleName

			points, err := service.FindConfigPoints(moduleFnCtx)
			if err != nil {
				return fmt.Errorf("finding config points: %w", err)
			}

			configPoints = make([]templates.ConfigPoint, len(points))
			for i, point := range points {
				configPoints[i] = templates.ConfigPoint{
					ModuleName:  point.ModuleName,
					Identifier:  point.Identifier,
					Name:        point.Name,
					Description: point.Description,
				}
			}

			wm, err := a.FindWingman(ctx, user.Username, environment.Slug, service.ModuleName, service.ServiceName)
			if err != nil {
				return fmt.Errorf("getting wingman: %w", err)
			}

			if wm != nil {
				cfp, err := wm.GetConfigPoints(ctx, service.ModuleName, state)
				if err != nil {
					return fmt.Errorf("getting wingman config points: %w", err)
				}

				for _, point := range cfp {
					configPoints = append(configPoints, templates.ConfigPoint{
						ModuleName:  service.ModuleName,
						Identifier:  point.Identifier,
						Name:        point.Name,
						Description: point.Description,
					})
				}
			}
		}
	}

	return templates.Deployment(deployment.Id, configPoints).Render(ctx, w)
}
