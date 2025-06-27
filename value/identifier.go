package value

import "fmt"

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

func AsIdentifier(value Value) (string, error) {
	if value.Kind() != Identifier {
		return "", fmt.Errorf("expected identifier, found %s", value.Kind())
	}

	return value.String(), nil
}
