package main

import (
	"embed"
	"fmt"
	"net/http"

	"github.com/BSFishy/mora-manager/router"
	"github.com/BSFishy/mora-manager/templates"
	"github.com/a-h/templ"
	"k8s.io/client-go/kubernetes"
)

//go:embed all:assets
var assets embed.FS

type App struct {
	clientset *kubernetes.Clientset
}

func NewApp() App {
	clientset, err := NewClientset()
	if err != nil {
		panic(err)
	}

	return App{
		clientset: clientset,
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

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("location", "/secret")
		w.WriteHeader(http.StatusTemporaryRedirect)
	})

	r.HandleGet("/secret", templ.Handler(templates.Secret()))
	r.Prefix("/assets", http.FileServerFS(assets))

	r.ListenAndServe(":8080")
}
