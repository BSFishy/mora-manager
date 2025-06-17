package value

type IdentifierValue struct {
	value string
}

func NewIdentifier(value string) Value {
	return IdentifierValue{
		value: value,
	}
}

func (i IdentifierValue) Kind() Kind {
	return Identifier
}

func (i IdentifierValue) String() string {
	return i.value
}

func (i IdentifierValue) Boolean() bool {
	return false
}

func (i IdentifierValue) Integer() int {
	return 0
}
