package value

type IntegerValue struct {
	value int
}

func NewInteger(value int) Value {
	return IntegerValue{
		value: value,
	}
}

func (i IntegerValue) Kind() Kind {
	return Integer
}

func (i IntegerValue) String() string {
	return ""
}

func (i IntegerValue) Boolean() bool {
	return false
}

func (i IntegerValue) Integer() int {
	return i.value
}

func (i IntegerValue) MarshalJSON() ([]byte, error) {
	return marshal(i.Kind(), i.value)
}
