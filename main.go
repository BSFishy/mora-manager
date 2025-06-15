package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/BSFishy/mora-manager/model"
	"github.com/BSFishy/mora-manager/router"
	"github.com/BSFishy/mora-manager/state"
	"github.com/BSFishy/mora-manager/templates"
	"github.com/a-h/templ"
	"k8s.io/client-go/kubernetes"
)

//go:embed all:assets
var assets embed.FS

type App struct {
	clientset *kubernetes.Clientset
	db        *model.DB
	secret    string
	registry  *FunctionRegistry
}

func NewApp() App {
	ctx := context.Background()

	clientset, err := NewClientset()
	if err != nil {
		panic(err)
	}

	db, err := model.NewDB()
	if err != nil {
		panic(err)
	}

	err = db.SetupMigrations(ctx)
	if err != nil {
		panic(err)
	}

	usersExist, err := db.UsersExist(ctx)
	if err != nil {
		panic(err)
	}

	var secret string
	if !usersExist {
		secret, err = db.GetOrCreateSecret(ctx)
		if err != nil {
			panic(err)
		}

		slog.Info("setup secret", "secret", secret)
	}

	registry := NewFunctionRegistry()
	RegisterConfigFunction(registry)

	return App{
		clientset: clientset,
		db:        db,
		secret:    secret,
		registry:  registry,
	}
}

func main() {
	SetupLogger()

	app := NewApp()

	r := router.NewRouter()

	r.RouteFunc("/api", func(r *router.Router) {
		r.RouteFunc("/v1", func(r *router.Router) {
			r.Use(app.apiMiddleware).Get("/ping", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, "pong")
			})

			r.Use(app.apiMiddleware).HandlePost("/image", router.ErrorHandlerFunc(app.imagePush))

			r.RouteFunc("/environment", func(r *router.Router) {
				r.RouteFunc("/:slug", func(r *router.Router) {
					r.Use(app.apiMiddleware).HandlePost("/deployment", router.ErrorHandlerFunc(app.createDeployment))
				})
			})
		})
	})

	// TODO: these middleware should probably set the htmx location headers for
	// redirects i think? unless the location header actually just redirects them.
	// in that case, it's fine.
	r.RouteFunc("/htmx", func(r *router.Router) {
		r.Use(app.secretMiddleware).Post("/secret", app.secretHtmxRoute)
		r.Use(app.userMiddleware).HandlePost("/user", router.ErrorHandlerFunc(app.userHtmxRoute))

		r.Use(app.loginMiddleware).HandlePost("/login", router.ErrorHandlerFunc(app.loginHtmxRoute))
		r.Use(app.userProtected).HandlePost("/signout", router.ErrorHandlerFunc(app.signOut))

		r.RouteFunc("/environment", func(r *router.Router) {
			r.Use(app.userProtected).HandlePost("/", router.ErrorHandlerFunc(app.createEnvironmentHtmxRoute))
			r.Use(app.userProtected).HandleDelete("/", router.ErrorHandlerFunc(app.deleteEnvironmentHtmxRoute))
		})

		r.RouteFunc("/deployment", func(r *router.Router) {
			r.Use(app.userProtected).HandleGet("/", router.ErrorHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				ctx := r.Context()
				user, _ := GetUser(ctx)

				environments, err := app.db.GetUserEnvironments(ctx, user.Id)
				if err != nil {
					return fmt.Errorf("getting environments: %w", err)
				}

				deployments, err := app.db.GetDeployments(ctx, environments)
				if err != nil {
					return fmt.Errorf("getting deployments: %w", err)
				}

				return templates.DashboardDeployments(environments, deployments).Render(ctx, w)
			}))

			r.Use(app.userProtected).HandlePost("/:id/config", router.ErrorHandlerFunc(app.updateDeploymentConfigHtmxRoute))
		})

		r.RouteFunc("/token", func(r *router.Router) {
			r.Use(app.userProtected).HandlePost("/", router.ErrorHandlerFunc(app.tokenHtmxRoute))
			r.Use(app.userProtected).HandlePost("/revoke", router.ErrorHandlerFunc(app.revokeTokenHtmxRoute))
		})
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/setup/secret", http.StatusFound)
	})

	r.Use(app.loginMiddleware).HandleGet("/login", templ.Handler(templates.Login()))
	r.Use(app.userProtected).HandleGet("/dashboard", router.ErrorHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		user, _ := GetUser(ctx)

		environments, err := app.db.GetUserEnvironments(ctx, user.Id)
		if err != nil {
			return fmt.Errorf("getting environments: %w", err)
		}

		deployments, err := app.db.GetDeployments(ctx, environments)
		if err != nil {
			return fmt.Errorf("getting deployments: %w", err)
		}

		return templates.Dashboard(templates.DashboardProps{
			User:         user,
			Environments: environments,
			Deployments:  deployments,
		}).Render(ctx, w)
	}))

	r.Use(app.userProtected).HandleGet("/deployment/:id", router.ErrorHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		user, _ := GetUser(ctx)

		params := router.Params(r)
		id := params["id"]

		deployment, err := app.db.GetDeployment(ctx, id)
		if err != nil {
			return fmt.Errorf("getting deployment: %w", err)
		}

		if deployment == nil {
			http.NotFound(w, r)
			return nil
		}

		environment, err := app.db.GetEnvironment(ctx, deployment.EnvironmentId)
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

			var state state.State
			if deployment.State != nil {
				if err = json.Unmarshal(*deployment.State, &state); err != nil {
					return fmt.Errorf("decoding state: %w", err)
				}
			}

			fnCtx := FunctionContext{
				Registry: app.registry,
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

				wm, err := app.FindWingman(ctx, user.Username, environment.Slug, service.ModuleName, service.ServiceName)
				if err != nil {
					return fmt.Errorf("getting wingman: %w", err)
				}

				if wm != nil {
					cfp, err := wm.GetConfigPoints(ctx, state)
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
	}))

	r.Use(app.userProtected).HandleGet("/environment", templ.Handler(templates.CreateEnvironment()))
	r.Use(app.userProtected).HandleGet("/tokens", router.ErrorHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		user, _ := GetUser(ctx)

		tokens, err := app.getTokenIds(ctx, user.Id)
		if err != nil {
			return fmt.Errorf("getting tokens: %w", err)
		}

		return templates.Tokens(templates.TokensProps{
			Tokens: tokens,
		}).Render(ctx, w)
	}))

	r.RouteFunc("/setup", func(r *router.Router) {
		r.Use(app.secretMiddleware).HandleGet("/secret", templ.Handler(templates.Secret()))
		r.Use(app.userMiddleware).HandleGet("/user", templ.Handler(templates.User()))
	})

	r.Prefix("/assets", http.FileServerFS(assets))

	r.ListenAndServe(":8080")
}
