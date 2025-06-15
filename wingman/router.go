package wingman

import (
	"github.com/BSFishy/mora-manager/router"
)

type app struct {
	wingman Wingman
}

func Start(wingman Wingman) {
	a := app{
		wingman: wingman,
	}

	r := router.NewRouter()

	r.HandlePost("/api/v1/config-point", router.ErrorHandlerFunc(a.handleConfigPoints))

	r.ListenAndServe(":8080")
}
