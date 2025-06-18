package main

import (
	"fmt"
	"net/http"

	"github.com/BSFishy/mora-manager/model"
	"github.com/BSFishy/mora-manager/templates"
)

func (a *App) dashboardPage(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user, _ := model.GetUser(ctx)

	environments, err := a.db.GetUserEnvironments(ctx, user.Id)
	if err != nil {
		return fmt.Errorf("getting environments: %w", err)
	}

	deployments, err := a.db.GetDeployments(ctx, environments)
	if err != nil {
		return fmt.Errorf("getting deployments: %w", err)
	}

	return templates.Dashboard(templates.DashboardProps{
		User:         user,
		Environments: environments,
		Deployments:  deployments,
	}).Render(ctx, w)
}
