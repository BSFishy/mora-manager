package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/BSFishy/mora-manager/logging"
)

type DeploymentRequest struct {
	Modules []Module `json:"modules"`
}

func (a *App) createDeployment(w http.ResponseWriter, req *http.Request) {
	var body DeploymentRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := req.Context()
	logger := logging.LogFromCtx(ctx)

	for _, module := range body.Modules {
		moduleLogger := logger.With(slog.Group("module", "name", module.Name))

		for _, service := range module.Services {
			serviceLogger := moduleLogger.With(slog.Group("service", "name", service.Name))

			serviceLogger.Info("deploying service")

			err := service.Deploy(ctx, a.clientset, module.Name)
			if err != nil {
				serviceLogger.Error("failed to deploy service", "err", err)
			}
		}
	}

	fmt.Fprint(w, "pong")
}
