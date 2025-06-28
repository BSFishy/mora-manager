package state

import "github.com/BSFishy/mora-manager/config"

type State struct {
	Configs      []StateConfig
	ServiceIndex int
}

// TODO: just use value.ServiceReferenceValue?
type ServiceRef struct {
	Module  string
	Service string
}

type StateConfig struct {
	ModuleName string
	Name       string
	Kind       config.PointKind
	Value      []byte
}

func (s *State) FindConfig(moduleName, name string) *StateConfig {
	for _, config := range s.Configs {
		if config.ModuleName == moduleName && config.Name == name {
			return &config
		}
	}

	return nil
}
