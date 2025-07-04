package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/BSFishy/mora-manager/api"
	"github.com/BSFishy/mora-manager/config"
	"github.com/BSFishy/mora-manager/kube"
	"github.com/BSFishy/mora-manager/model"
	"github.com/BSFishy/mora-manager/point"
	"github.com/BSFishy/mora-manager/router"
	statepkg "github.com/BSFishy/mora-manager/state"
	"github.com/BSFishy/mora-manager/templates"
	"github.com/BSFishy/mora-manager/util"
	v1 "k8s.io/api/core/v1"
)

type DeploymentResponse struct {
	Id string `json:"id"`
}

func (a *App) createDeployment(w http.ResponseWriter, req *http.Request) error {
	var cfg api.Config
	if err := json.NewDecoder(req.Body).Decode(&cfg); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	ctx := req.Context()
	user, _ := model.GetUser(ctx)

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

	// at this point, we shouldnt be actually referencing any configuration or
	// state related things, so this should be fine even though the state and
	// config structures are not in the context. might want to look into just
	// returning empty values for these for safety?
	services, err := config.ServiceConfigFromModules(ctx, a, cfg.Modules)
	if err != nil {
		return fmt.Errorf("sorting services: %w", err)
	}

	configs, err := cfg.FlattenConfigs(ctx, a)
	if err != nil {
		return fmt.Errorf("flattening configs: %w", err)
	}

	// TODO: i should be able to configure whether i want to inherit the previous
	// deployment's state
	previousDeployment, err := environment.GetLastDeployment(ctx, a.db)
	if err != nil {
		return fmt.Errorf("getting previous deployment: %w", err)
	}

	var previousDeploymentId *string
	if previousDeployment != nil {
		previousDeploymentId = &previousDeployment.Id
	}

	deployment, err := environment.NewDeployment(ctx, a.db, previousDeploymentId, config.Config{
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
	user, _ := model.GetUser(ctx)

	pageStr := r.URL.Query().Get("page")
	page := 1
	if pageStr != "" {
		if pageNum, err := strconv.Atoi(pageStr); err == nil {
			page = pageNum
		}
	}

	environments, err := a.db.GetUserEnvironments(ctx, user.Id)
	if err != nil {
		return fmt.Errorf("getting environments: %w", err)
	}

	totalPages, err := a.db.CountDeploymentPages(ctx, environments)
	if err != nil {
		return fmt.Errorf("counting deployments: %w", err)
	}

	deployments, err := a.db.GetDeployments(ctx, environments, page-1)
	if err != nil {
		return fmt.Errorf("getting deployments: %w", err)
	}

	return templates.DashboardDeployments(environments, deployments, totalPages, page).Render(ctx, w)
}

func (a *App) deploymentStatusHtmxRoute(w http.ResponseWriter, r *http.Request) error {
	props, err := a.getDeploymentProps(w, r)
	if err != nil {
		return err
	}

	if props == nil {
		return nil
	}

	return templates.DeploymentHtmx(*props).Render(r.Context(), w)
}

func (a *App) updateDeploymentConfigHtmxRoute(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	logger := util.LogFromCtx(ctx)

	user, _ := model.GetUser(ctx)

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

	var previousDeployment *model.Deployment
	if d.PreviousDeploymentId != nil {
		previousDeployment, err = a.db.GetDeployment(ctx, *d.PreviousDeploymentId)
		if err != nil {
			return fmt.Errorf("getting previous deployment: %w", err)
		}
	}

	moduleNames := r.Form["module_name"]
	identifiers := r.Form["identifier"]
	values := r.Form["value"]
	inherits := r.Form["inherit"]

	if len(moduleNames) != len(identifiers) || len(moduleNames) != len(values) || len(moduleNames) != len(inherits) {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	err = a.db.Transact(ctx, func(tx *sql.Tx) error {
		err := d.Lock(ctx, tx)
		if err != nil {
			return fmt.Errorf("taking deployment lock: %w", err)
		}

		var cfg config.Config
		if err := json.Unmarshal(d.Config, &cfg); err != nil {
			return fmt.Errorf("decoding config: %w", err)
		}

		var state statepkg.State
		if d.State != nil {
			if err := json.Unmarshal(*d.State, &state); err != nil {
				return fmt.Errorf("decoding state: %w", err)
			}
		}

		var previousState statepkg.State
		if previousDeployment != nil && previousDeployment.State != nil {
			if err := json.Unmarshal(*previousDeployment.State, &previousState); err != nil {
				return fmt.Errorf("decoding previous state: %w", err)
			}
		}

		for i := range moduleNames {
			moduleName := moduleNames[i]
			identifier := identifiers[i]
			v := []byte(values[i])
			inherit := inherits[i]

			services := cfg.Services[state.ServiceIndex:]
			if len(services) < 1 {
				return fmt.Errorf("invalid service to configure: %d", state.ServiceIndex)
			}

			service := services[0]
			if service.ModuleName != moduleName {
				w.WriteHeader(http.StatusBadRequest)
				return nil
			}

			runwayCtx := &runwayContext{
				manager:     nil,
				clientset:   a.clientset,
				registry:    a.registry,
				user:        user,
				environment: env,
				config:      &cfg,
				state:       &state,
				moduleName:  moduleName,
				serviceName: service.ServiceName,
			}

			cfps, err := service.FindConfigPoints(ctx, runwayCtx)
			if err != nil {
				return fmt.Errorf("finding config points: %w", err)
			}

			c := cfps.Find(moduleName, identifier)
			if c == nil {
				w.WriteHeader(http.StatusBadRequest)
				return nil
			}

			if inherit == "true" {
				previousValue := previousState.FindConfig(moduleName, identifier)
				if previousValue == nil {
					return fmt.Errorf("failed to find %s/%s in previous config", moduleName, identifier)
				}

				logger.Debug("inheriting config from previous deployment", "moduleName", moduleName, "identifier", identifier)
				state.Configs = append(state.Configs, *previousValue)
				continue
			}

			if c.Kind == point.Secret {
				secret := kube.NewSecret(runwayCtx, identifier, v)
				deployment := kube.MaterializedService{
					Secrets: []kube.Resource[v1.Secret]{secret},
				}

				if err = deployment.Deploy(ctx, runwayCtx); err != nil {
					return err
				}

				state.Configs = append(state.Configs, statepkg.StateConfig{
					ModuleName: moduleName,
					Name:       identifier,
					Kind:       point.Secret,
					Value:      []byte(secret.Name()),
				})
			} else {
				state.Configs = append(state.Configs, statepkg.StateConfig{
					ModuleName: moduleName,
					Name:       identifier,
					Kind:       point.String,
					Value:      v,
				})
			}
		}

		if err := d.UpdateStateAndStatus(ctx, tx, model.InProgress, state); err != nil {
			return fmt.Errorf("updating state: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	go a.deploy(d)

	props, err := a.getDeploymentProps(w, r)
	if err != nil {
		return fmt.Errorf("getting props: %w", err)
	}

	if props == nil {
		return nil
	}

	return templates.DeploymentHtmx(*props).Render(r.Context(), w)
}

func (a *App) getDeploymentProps(w http.ResponseWriter, r *http.Request) (*templates.DeploymentProps, error) {
	ctx := r.Context()
	user, _ := model.GetUser(ctx)

	params := router.Params(r)
	id := params["id"]

	deployment, err := a.db.GetDeployment(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting deployment: %w", err)
	}

	if deployment == nil {
		http.NotFound(w, r)
		return nil, nil
	}

	environment, err := a.db.GetEnvironment(ctx, deployment.EnvironmentId)
	if err != nil {
		return nil, fmt.Errorf("getting environment: %w", err)
	}

	var previousDeployment *model.Deployment
	if deployment.PreviousDeploymentId != nil {
		previousDeployment, err = a.db.GetDeployment(ctx, *deployment.PreviousDeploymentId)
		if err != nil {
			return nil, fmt.Errorf("getting previous deployment: %w", err)
		}
	}

	if environment.UserId != user.Id {
		http.NotFound(w, r)
		return nil, nil
	}

	var configPoints []point.Point
	var values []string
	if deployment.Status == model.Waiting {
		var cfg config.Config
		if err = json.Unmarshal(deployment.Config, &cfg); err != nil {
			return nil, fmt.Errorf("decoding config: %w", err)
		}

		var state statepkg.State
		if deployment.State != nil {
			if err = json.Unmarshal(*deployment.State, &state); err != nil {
				return nil, fmt.Errorf("decoding state: %w", err)
			}
		}

		var previousState statepkg.State
		if previousDeployment != nil && previousDeployment.State != nil {
			if err = json.Unmarshal(*previousDeployment.State, &previousState); err != nil {
				return nil, fmt.Errorf("decoding previous state: %w", err)
			}
		}

		services := cfg.Services[state.ServiceIndex:]
		if len(services) > 0 {
			service := services[0]

			runwayCtx := &runwayContext{
				manager:     nil,
				clientset:   a.clientset,
				registry:    a.registry,
				user:        user,
				environment: environment,
				config:      &cfg,
				state:       &state,
				moduleName:  service.ModuleName,
				serviceName: service.ServiceName,
			}

			configPoints, err = service.FindConfigPoints(ctx, runwayCtx)
			if err != nil {
				return nil, fmt.Errorf("finding config points: %w", err)
			}

			for _, p := range configPoints {
				stateValue := previousState.FindConfig(p.ModuleName, p.Identifier)
				if stateValue != nil {
					if stateValue.Kind == point.Secret {
						values = append(values, "asdf")
					} else {
						values = append(values, string(stateValue.Value))
					}
				} else {
					values = append(values, "")
				}
			}
		}
	}

	return &templates.DeploymentProps{
		Id:           deployment.Id,
		Status:       deployment.Status,
		ConfigPoints: configPoints,
		Values:       values,
	}, nil
}

func (a *App) deploymentPage(w http.ResponseWriter, r *http.Request) error {
	props, err := a.getDeploymentProps(w, r)
	if err != nil {
		return err
	}

	if props == nil {
		return nil
	}

	return templates.Deployment(*props).Render(r.Context(), w)
}
