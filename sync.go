package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type SyncRequest struct {
	Modules []Module `json:"modules"`
}

func (a *App) Sync(w http.ResponseWriter, req *http.Request) {
	var body SyncRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := req.Context()

	for _, module := range body.Modules {
		for _, service := range module.Services {
			fmt.Printf("deploying service %s/%s\n", module.Name, service.Name)

			err := service.Deploy(ctx, a.clientset, module.Name)
			if err != nil {
				fmt.Printf("failed to deploy service: %s\n", err)
			}
		}
	}

	fmt.Fprint(w, "pong")
}
