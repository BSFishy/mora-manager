package main

import (
	"fmt"
	"net/http"
)

func main() {
	r := NewRouter()
	r = *r.Use(log)

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "pong")
	})

	r.Post("/sync", Sync)

	r.ListenAndServe(":8080")
}
