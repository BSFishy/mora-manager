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

type Value interface {
	Kind() Kind

	String() string
	Boolean() bool
	Integer() int
}
