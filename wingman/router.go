package wingman

import (
	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/router"
)

type app struct {
	wingman  Wingman
	registry *expr.FunctionRegistry
}

func Start(wingman Wingman) {
	// TODO: fill this with default functions
	registry := expr.NewFunctionRegistry()

	a := app{
		wingman:  wingman,
		registry: registry,
	}

	r := router.NewRouter()

	r.HandlePost("/api/v1/config-point", router.ErrorHandlerFunc(a.handleConfigPoints))
	r.HandlePost("/api/v1/function", router.ErrorHandlerFunc(a.handleFunction))

	r.ListenAndServe(":8080")
}
