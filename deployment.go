package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/model"
	"github.com/BSFishy/mora-manager/router"
	statepkg "github.com/BSFishy/mora-manager/state"
	"github.com/BSFishy/mora-manager/templates"
	"github.com/BSFishy/mora-manager/util"
	"github.com/BSFishy/mora-manager/value"
)

type ApiConfig struct {
	Modules []Module `json:"modules"`
}

type Module struct {
	Name     string         `json:"name"`
	Services []Service      `json:"services"`
	Configs  []ModuleConfig `json:"configs"`
}

type ModuleConfig struct {
	// this field doesnt exist on the api but it is useful for passing these
	// around
	ModuleName  string
	Identifier  string
	Name        expr.Expression
	Description *expr.Expression
}

type ApiWingman struct {
	Image expr.Expression
}

type Service struct {
	Name     string            `json:"name"`
	Image    expr.Expression   `json:"image"`
	Requires []expr.Expression `json:"requires"`
	Wingman  *ApiWingman       `json:"wingman,omitempty"`
}

func (s *Service) RequiredServices(ctx context.Context) ([]statepkg.ServiceRef, error) {
	services := []statepkg.ServiceRef{}
	for _, service := range s.Requires {
		v, cfp, err := service.Evaluate(ctx)
		if err != nil {
			return nil, err
		}

		if len(cfp) > 0 {
			return nil, errors.New("unexpected configurable value")
		}

		ref, ok := v.(value.ServiceReferenceValue)
		if !ok {
			return nil, errors.New("invalid requires value")
		}

		// TODO: can i just use value.ServiceReferenceValue here?
		services = append(services, statepkg.ServiceRef{
			Module:  ref.ModuleName,
			Service: ref.ServiceName,
		})
	}

	return services, nil
}

func (c *ApiConfig) FlattenConfigs() []ModuleConfig {
	configs := []ModuleConfig{}
	for _, module := range c.Modules {
		for _, config := range module.Configs {
			config.ModuleName = module.Name
			configs = append(configs, config)
		}
	}

	return configs
}

func (c *ApiConfig) TopologicalSort(ctx context.Context) ([]ServiceConfig, error) {
	services := make(map[string]ServiceConfig)
	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	for _, module := range c.Modules {
		for _, service := range module.Services {
			path := fmt.Sprintf("%s/%s", module.Name, service.Name)
			requires, err := service.RequiredServices(ctx)
			if err != nil {
				return nil, fmt.Errorf("getting required services: %w", err)
			}

			var wingman *ServiceWingman
			if service.Wingman != nil {
				wingman = &ServiceWingman{
					Image: service.Wingman.Image,
				}
			}

			services[path] = ServiceConfig{
				ModuleName:  module.Name,
				ServiceName: service.Name,
				Image:       service.Image,
				Wingman:     wingman,
			}

			for _, dep := range requires {
				depPath := fmt.Sprintf("%s/%s", dep.Module, dep.Service)
				graph[depPath] = append(graph[depPath], path)
				inDegree[path]++
			}

			if _, ok := inDegree[path]; !ok {
				inDegree[path] = 0
			}
		}
	}

	var queue []string
	for path, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, path)
		}
	}

	var result []ServiceConfig
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		result = append(result, services[cur])

		for _, neighbor := range graph[cur] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	return result, nil
}

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

	ctx = expr.WithFunctionRegistry(ctx, a.registry)

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

	ctx = WithEnvironment(ctx, environment)

	if err = environment.CancelInProgressDeployments(ctx, a.db); err != nil {
		return fmt.Errorf("cancelling deployments: %w", err)
	}

	// at this point, we shouldnt be actually referencing any configuration or
	// state related things, so this should be fine even though the state and
	// config structures are not in the context. might want to look into just
	// returning empty values for these for safety?
	services, err := config.TopologicalSort(ctx)
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
			v := values[i]

			if cfg := config.FindConfig(moduleName, identifier); cfg == nil {
				w.WriteHeader(http.StatusBadRequest)
				return nil
			}

			state.Configs = append(state.Configs, statepkg.StateConfig{
				ModuleName: moduleName,
				Name:       identifier,
				Value:      v,
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

	ctx = expr.WithFunctionRegistry(ctx, a.registry)

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

	ctx = WithEnvironment(ctx, environment)

	if environment.UserId != user.Id {
		http.NotFound(w, r)
		return nil
	}

	var configPoints []value.ConfigPoint
	if deployment.Status == model.Waiting {
		var config Config
		if err = json.Unmarshal(deployment.Config, &config); err != nil {
			return fmt.Errorf("decoding config: %w", err)
		}

		ctx = WithConfig(ctx, &config)

		var state statepkg.State
		if deployment.State != nil {
			if err = json.Unmarshal(*deployment.State, &state); err != nil {
				return fmt.Errorf("decoding state: %w", err)
			}
		}

		ctx = WithState(ctx, &state)

		services := config.Services[state.ServiceIndex:]
		if len(services) > 0 {
			service := services[0]
			ctx := WithModuleName(ctx, service.ModuleName)
			ctx = WithServiceName(ctx, service.ServiceName)

			_, cfp, err := service.Evaluate(ctx)
			if err != nil {
				return fmt.Errorf("evaluating service: %w", err)
			}

			configPoints = append(configPoints, cfp...)

			wm, err := a.FindWingman(ctx)
			if err != nil {
				return fmt.Errorf("getting wingman: %w", err)
			}

			if wm != nil {
				cfp, err := wm.GetConfigPoints(ctx)
				if err != nil {
					return fmt.Errorf("getting wingman config points: %w", err)
				}

				for _, point := range cfp {
					// TODO: ideally these are not separate types
					configPoints = append(configPoints, value.ConfigPoint{
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
