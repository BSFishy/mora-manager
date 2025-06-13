package main

type State struct {
	Configs          []StateConfig
	DeployedServices []ServiceRef
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

func (s *State) AddDeployedService(moduleName, name string) {
	s.DeployedServices = append(s.DeployedServices, ServiceRef{
		Module:  moduleName,
		Service: name,
	})
}

func (s *State) FilterDeployedServices(services []ServiceConfig) []ServiceConfig {
	result := []ServiceConfig{}
	for _, service := range services {
		found := false
		for _, ref := range s.DeployedServices {
			if service.ModuleName == ref.Module && service.ServiceName == ref.Service {
				found = true
				break
			}
		}

		if found {
			continue
		}

		result = append(result, service)
	}

	return result
}
