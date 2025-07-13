package value

type Kind string

const (
	Null             Kind = "null"
	Raw              Kind = "raw"
	String           Kind = "string"
	Identifier       Kind = "identifier"
	Secret           Kind = "secret"
	Boolean          Kind = "boolean"
	Integer          Kind = "integer"
	ServiceReference Kind = "service_reference"
)

func (k Kind) String() string {
	return string(k)
}

type Value interface {
	Kind() Kind

	Raw() []byte
	String() string
	Boolean() bool
	Integer() int
}
