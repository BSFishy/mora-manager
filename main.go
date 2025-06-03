package main

import (
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
}

func NewApp() App {
	clientset, err := NewClientset()
	if err != nil {
		panic(err)
	}

	db, err := model.NewDB()
	if err != nil {
		panic(err)
	}

	err = db.SetupMigrations()
	if err != nil {
		panic(err)
	}

	usersExist, err := db.UsersExist()
	if err != nil {
		panic(err)
	}

	var secret string
	if !usersExist {
		secret, err = db.GetOrCreateSecret()
		if err != nil {
			panic(err)
		}

		slog.Info("setup secret", "secret", secret)
	}

	return App{
		clientset: clientset,
		db:        db,
		secret:    secret,
	}
}

func main() {
	SetupLogger()

	app := NewApp()

	r := router.NewRouter()
	r = *r.Use(log)

	r.RouteFunc("/api", func(r *router.Router) {
		r.RouteFunc("/v1", func(r *router.Router) {
			r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, "pong")
			})

			r.Post("/deployment", app.createDeployment)
		})
	})

	// TODO: gate this function with if users exist
	r.RouteFunc("/htmx", func(r *router.Router) {
		r.Post("/secret", func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := LogFromCtx(ctx)

			err := r.ParseForm()
			if err != nil {
				logger.Error("failed to parse secret form", "err", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if !r.Form.Has("secret") {
				templates.SecretForm(true).Render(ctx, w)
				return
			}

			secret := r.Form.Get("secret")
			if secret != app.secret {
				templates.SecretForm(true).Render(ctx, w)
				return
			}

			// TODO: create a session, send it as a cookie, redirect to admin setup
			fmt.Fprint(w, "pong")
		})
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("location", "/secret")
		w.WriteHeader(http.StatusTemporaryRedirect)
	})

	// TODO: gate this route with if users exist
	r.HandleGet("/secret", templ.Handler(templates.Secret()))
	r.Prefix("/assets", http.FileServerFS(assets))

	r.ListenAndServe(":8080")
}
