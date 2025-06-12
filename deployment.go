package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/BSFishy/mora-manager/router"
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
		w.WriteHeader(http.StatusNotFound)
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
