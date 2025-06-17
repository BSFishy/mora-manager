package state

type State struct {
	Configs      []StateConfig
	ServiceIndex int
}

// TODO: just use value.ServiceReferenceValue?
type ServiceRef struct {
	Module  string
	Service string
}

// TODO: this structure sucks. it needs to be changed to something more useful
type StateConfig struct {
	ModuleName string
	Name       string
	Value      string
}

func (s *State) FindConfig(moduleName, name string) *StateConfig {
	for _, config := range s.Configs {
		if config.ModuleName == moduleName && config.Name == name {
			return &config
		}
	}

	return nil
}
