package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/BSFishy/mora-manager/model"
	"github.com/BSFishy/mora-manager/templates"
)

func (a *App) dashboardPage(w http.ResponseWriter, r *http.Request) error {
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

	return templates.Dashboard(templates.DashboardProps{
		User:         user,
		Environments: environments,
		Deployments:  deployments,
		TotalPages:   totalPages,
		Page:         page,
	}).Render(ctx, w)
}
