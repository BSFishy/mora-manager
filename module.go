package main

type Module struct {
	Name     string    `json:"name"`
	Services []Service `json:"services"`
}

type Service struct {
	Name  string     `json:"name"`
	Image Expression `json:"image"`
}

type Expression struct {
	Atom *Atom         `json:"atom,omitempty"`
	List *[]Expression `json:"list,omitempty"`
}

type Atom struct {
	Identifier *string `json:"identifier,omitempty"`
	String     *string `json:"string,omitempty"`
	Number     *string `json:"number,omitempty"`
}
