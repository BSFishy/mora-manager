package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/BSFishy/mora-manager/model"
	"github.com/BSFishy/mora-manager/templates"
)

func (a *App) createEnvironmentHtmxRoute(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user, _ := model.GetUser(ctx)

	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parsing form: %w", err)
	}

	name := r.Form.Get("name")
	if name == "" {
		return templates.CreateEnvironmentForm(true).Render(ctx, w)
	}

	slug := strings.ToLower(r.Form.Get("slug"))
	if slug == "" {
		return templates.CreateEnvironmentForm(true).Render(ctx, w)
	}

	environment, err := a.db.GetEnvironmentBySlug(ctx, user.Id, slug)
	if err != nil {
		return fmt.Errorf("getting environment: %w", err)
	}

	if environment != nil {
		return templates.CreateEnvironmentForm(true).Render(ctx, w)
	}

	_, err = a.db.NewEnvironment(ctx, user.Id, name, slug)
	if err != nil {
		return fmt.Errorf("creating environment: %w", err)
	}

	w.Header().Set("Hx-Location", "/dashboard")
	return nil
}

func (a *App) deleteEnvironmentHtmxRoute(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user, _ := model.GetUser(ctx)

	environmentId := r.URL.Query().Get("id")
	if environmentId == "" {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	environment, err := a.db.GetEnvironment(ctx, environmentId)
	if err != nil {
		return fmt.Errorf("getting environment: %w", err)
	}

	if environment == nil || environment.UserId != user.Id {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	err = environment.Delete(ctx, a.db)
	if err != nil {
		return fmt.Errorf("deleting environment: %w", err)
	}

	environments, err := a.db.GetUserEnvironments(ctx, user.Id)
	if err != nil {
		return fmt.Errorf("getting user environments: %w", err)
	}

	return templates.DashboardEnvironments(environments).Render(ctx, w)
}
