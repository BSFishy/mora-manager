package wingman

import "context"

type ConfigPoint struct {
	Identifier  string
	Name        string
	Description *string
}

type Wingman interface {
	GetConfigPoints(context.Context) ([]ConfigPoint, error)
}
