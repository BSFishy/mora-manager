package config

import (
	"context"

	"github.com/BSFishy/mora-manager/util"
)

type PointKind string

const (
	String PointKind = "string"
	Secret PointKind = "secret"
)

type Point struct {
	ModuleName string
	// slug used to identify this point internally. must be module-unique
	Identifier string
	// pretty, human readable name
	Name string
	Kind PointKind
	// optional description of the point
	Description *string
}

// Fill infers values from the context to the point if they are empty
func (p *Point) Fill(ctx context.Context) {
	if p.ModuleName == "" {
		p.ModuleName = util.Has(util.GetModuleName(ctx))
	}

	if p.Kind == "" {
		p.Kind = String
	}
}

type Points []Point

func (p Points) Find(moduleName, identifier string) *Point {
	for _, point := range p {
		if point.ModuleName == moduleName && point.Identifier == identifier {
			return &point
		}
	}

	return nil
}
