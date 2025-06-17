package main

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/BSFishy/mora-manager/model"
	"github.com/BSFishy/mora-manager/router"
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
	RegisterDefaultFunctions(registry)

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
			r.Use(app.userProtected).HandleGet("/", router.ErrorHandlerFunc(app.deploymentHtmxRoute))
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
	r.Use(app.userProtected).HandleGet("/dashboard", router.ErrorHandlerFunc(app.dashboardPage))

	r.Use(app.userProtected).HandleGet("/deployment/:id", router.ErrorHandlerFunc(app.deploymentPage))
	r.Use(app.userProtected).HandleGet("/environment", templ.Handler(templates.CreateEnvironment()))
	r.Use(app.userProtected).HandleGet("/tokens", router.ErrorHandlerFunc(app.tokenPage))

	r.RouteFunc("/setup", func(r *router.Router) {
		r.Use(app.secretMiddleware).HandleGet("/secret", templ.Handler(templates.Secret()))
		r.Use(app.userMiddleware).HandleGet("/user", templ.Handler(templates.User()))
	})

	r.Prefix("/assets", http.FileServerFS(assets))

	r.ListenAndServe(":8080")
}
