package main

import (
	"fmt"
	"net/http"

	"k8s.io/client-go/kubernetes"
)

type App struct {
	clientset *kubernetes.Clientset
}

func main() {
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

			r.Post("/state", app.Sync)
		})
	})

	r.ListenAndServe(":8080")
}
