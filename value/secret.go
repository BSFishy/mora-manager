package value

type SecretValue struct {
	value string
}

func NewSecret(value string) Value {
	return SecretValue{
		value: value,
	}
}

func (s SecretValue) Kind() Kind {
	return Secret
}

func (s SecretValue) String() string {
	return s.value
}

func (s SecretValue) Boolean() bool {
	return false
}

func (s SecretValue) Integer() int {
	return 0
}

func (s SecretValue) MarshalJSON() ([]byte, error) {
	return marshal(s.Kind(), s.value)
}
