package main

import (
	"embed"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"k8s.io/client-go/kubernetes"
)

//go:embed all:assets
var assets embed.FS

type App struct {
	clientset *kubernetes.Clientset
}

func main() {
	SetupLogger()

	clientset, err := NewClientset()
	if err != nil {
		panic(err)
	}

	app := App{
		clientset: clientset,
	}

	r := NewRouter()
	r = *r.Use(log)

	r.RouteFunc("/api", func(r *Router) {
		r.RouteFunc("/v1", func(r *Router) {
			r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, "pong")
			})

			r.Post("/deployment", app.Sync)
		})
	})

	r.HandleGet("/", templ.Handler(index()))
	r.HandleGet("/assets/", http.FileServerFS(assets))

	r.ListenAndServe(":8080")
}
