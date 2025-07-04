package value

type BooleanValue struct {
	value bool
}

func NewBoolean(value bool) Value {
	return BooleanValue{
		value: value,
	}
}

func (b BooleanValue) Kind() Kind {
	return Boolean
}

func (b BooleanValue) String() string {
	return ""
}

func (b BooleanValue) Boolean() bool {
	return b.value
}

func (b BooleanValue) Integer() int {
	return 0
}

func (b BooleanValue) MarshalJSON() ([]byte, error) {
	return marshal(b.Kind(), b.value)
}
