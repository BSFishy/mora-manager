package main

import (
	"errors"
)

type Module struct {
	Name     string    `json:"name"`
	Services []Service `json:"services"`
}

type Service struct {
	Name     string       `json:"name"`
	Image    Expression   `json:"image"`
	Requires []Expression `json:"requires"`
}

type ServiceRef struct {
	Module  string
	Service string
}

func (s *Service) RequiredServices() ([]ServiceRef, error) {
	services := []ServiceRef{}
	for _, service := range s.Requires {
		list := service.List
		if list == nil {
			return nil, errors.New("required service is not a list")
		}

		expr := *list
		if len(expr) != 3 {
			return nil, errors.New("requires function call invalid")
		}

		f := expr[0].AsIdentifier()
		if f == nil {
			return nil, errors.New("list function is not an identifier")
		}

		if *f != "service" {
			return nil, errors.New("requires function is not service")
		}

		moduleName := expr[1].AsIdentifier()
		if moduleName == nil {
			return nil, errors.New("service reference module name is not an identifier")
		}

		serviceName := expr[2].AsIdentifier()
		if serviceName == nil {
			return nil, errors.New("service reference service name is not an identifier")
		}

		services = append(services, ServiceRef{
			Module:  *moduleName,
			Service: *serviceName,
		})
	}

	return services, nil
}

type Expression struct {
	Atom *Atom         `json:"atom,omitempty"`
	List *[]Expression `json:"list,omitempty"`
}

func (e *Expression) AsIdentifier() *string {
	if e.Atom == nil {
		return nil
	}

	return e.Atom.Identifier
}

func (e *Expression) AsString() *string {
	if e.Atom == nil {
		return nil
	}

	return e.Atom.String
}

type Atom struct {
	Identifier *string `json:"identifier,omitempty"`
	String     *string `json:"string,omitempty"`
	Number     *string `json:"number,omitempty"`
}
