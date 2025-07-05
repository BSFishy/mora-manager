package wingman

import (
	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/function"
	"github.com/BSFishy/mora-manager/kube"
	"github.com/BSFishy/mora-manager/router"
	"k8s.io/client-go/kubernetes"
)

type app struct {
	client   *kubernetes.Clientset
	wingman  Wingman
	registry expr.FunctionRegistry
}

func Start(wingman Wingman) {
	manager := &Manager{}
	registry := function.NewRegistry(manager)

	client, err := kube.NewClientset()
	if err != nil {
		panic(err)
	}

	a := app{
		client:   client,
		wingman:  wingman,
		registry: registry,
	}

	r := router.NewRouter()

	r.HandlePost("/api/v1/config-point", router.ErrorHandlerFunc(a.handleConfigPoints))
	r.HandlePost("/api/v1/function", router.ErrorHandlerFunc(a.handleFunction))

	r.ListenAndServe(":8080")
}
