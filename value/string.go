package value

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
