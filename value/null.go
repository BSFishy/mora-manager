package value

type NullValue struct{}

func NewNull() Value {
	return NullValue{}
}

func (n NullValue) Kind() Kind {
	return Null
}

func (n NullValue) String() string {
	return ""
}

func (n NullValue) Boolean() bool {
	return false
}

func (n NullValue) Integer() int {
	return 0
}
