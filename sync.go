package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type SyncRequest struct {
	Modules []Module `json:"modules"`
}

func Sync(w http.ResponseWriter, req *http.Request) {
	var body SyncRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, module := range body.Modules {
		for _, service := range module.Services {
			fmt.Printf("service %s/%s: %s\n", module.Name, service.Name, *service.Image.Atom.String)
		}
	}

	fmt.Fprint(w, "pong")
}
