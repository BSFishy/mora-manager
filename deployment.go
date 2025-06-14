package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/BSFishy/mora-manager/router"
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

	tx, err := d.Lock(ctx, a.db)
	if err != nil {
		return fmt.Errorf("taking deployment lock: %w", err)
	}

	defer d.Unlock(tx)

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

	var state State
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

		state.Configs = append(state.Configs, StateConfig{
			ModuleName: moduleName,
			Name:       identifier,
			Value:      value,
		})
	}

	if err := d.UpdateState(ctx, tx, state); err != nil {
		return fmt.Errorf("updating state: %w", err)
	}

	go a.deploy(d)

	return nil
}
