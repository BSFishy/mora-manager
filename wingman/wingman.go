package wingman

import (
	"context"

	"github.com/BSFishy/mora-manager/state"
)

type ConfigPoint struct {
	Identifier  string
	Name        string
	Description *string
}

type Wingman interface {
	GetConfigPoints(context.Context, state.State) ([]ConfigPoint, error)
}
