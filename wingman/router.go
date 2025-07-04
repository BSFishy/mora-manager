package wingman

import (
	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/function"
	"github.com/BSFishy/mora-manager/router"
)

type app struct {
	wingman  Wingman
	registry expr.FunctionRegistry
}

func Start(wingman Wingman) {
	manager := &Manager{}
	registry := function.NewRegistry(manager)

	a := app{
		wingman:  wingman,
		registry: registry,
	}

	r := router.NewRouter()

	r.HandlePost("/api/v1/config-point", router.ErrorHandlerFunc(a.handleConfigPoints))
	r.HandlePost("/api/v1/function", router.ErrorHandlerFunc(a.handleFunction))

	r.ListenAndServe(":8080")
}
