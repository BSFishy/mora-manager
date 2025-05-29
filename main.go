package main

import (
	"fmt"
	"net/http"
)

func main() {
	r := NewRouter()
	r = *r.Use(log)

	r.RouteFunc("/api", func(r *Router) {
		r.RouteFunc("/v1", func(r *Router) {
			r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, "pong")
			})

			r.Post("/state", Sync)
		})
	})

	r.ListenAndServe(":8080")
}
