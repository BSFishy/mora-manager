package state

type State struct {
	Configs      []StateConfig
	ServiceIndex int
}

type ServiceRef struct {
	Module  string
	Service string
}

type StateConfig struct {
	ModuleName string
	Name       string
	Value      any
}

func (s *State) FindConfig(moduleName, name string) *StateConfig {
	for _, config := range s.Configs {
		if config.ModuleName == moduleName && config.Name == name {
			return &config
		}
	}

	return nil
}
