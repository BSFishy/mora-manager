package value

import "fmt"

type RawValue struct {
	value []byte
}

func NewRaw(value []byte) Value {
	return RawValue{
		value: value,
	}
}

func (r RawValue) Kind() Kind {
	return Raw
}

func (r RawValue) Raw() []byte {
	return r.value
}

func (r RawValue) String() string {
	return string(r.value)
}

func (r RawValue) Boolean() bool {
	return false
}

func (r RawValue) Integer() int {
	return 0
}

func (r RawValue) MarshalJSON() ([]byte, error) {
	return marshal(Raw, r.value)
}

func AsRaw(value Value) ([]byte, error) {
	if value.Kind() != Raw {
		return nil, fmt.Errorf("expected raw, found %s", value.Kind())
	}

	return value.Raw(), nil
}
