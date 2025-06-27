package wingman

import (
	"context"

	"github.com/BSFishy/mora-manager/config"
)

type Wingman interface {
	GetConfigPoints(context.Context) ([]config.Point, error)
}
