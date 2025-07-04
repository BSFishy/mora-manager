package value

import (
	"fmt"
)

type StringValue struct {
	value string
}

func NewString(value string) Value {
	return StringValue{
		value: value,
	}
}

func (s StringValue) Kind() Kind {
	return String
}

func (s StringValue) String() string {
	return s.value
}

func (s StringValue) Boolean() bool {
	return false
}

func (s StringValue) Integer() int {
	return 0
}

func (s StringValue) MarshalJSON() ([]byte, error) {
	return marshal(s.Kind(), s.value)
}

func AsString(value Value) (string, error) {
	if value.Kind() != String {
		return "", fmt.Errorf("expected string, found %s", value.Kind())
	}

	return value.String(), nil
}
