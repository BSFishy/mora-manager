package value

type Kind int

const (
	Null Kind = iota
	String
	Identifier
	Secret
	Boolean
	Integer
	ServiceReference
)

func (k Kind) String() string {
	switch k {
	case Null:
		return "null"
	case String:
		return "string"
	case Identifier:
		return "identifier"
	case Secret:
		return "secret"
	case Boolean:
		return "boolean"
	case Integer:
		return "integer"
	case ServiceReference:
		return "service reference"
	}

	panic("unknown value kind")
}

type Value interface {
	Kind() Kind

	String() string
	Boolean() bool
	Integer() int
}
