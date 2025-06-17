package value

// this is a special value and needs to be type-casted to get out the relevant
// information
type ServiceReferenceValue struct {
	ModuleName  string
	ServiceName string
}

func NewServiceReference(moduleName, serviceName string) Value {
	return ServiceReferenceValue{
		ModuleName:  moduleName,
		ServiceName: serviceName,
	}
}

func (s ServiceReferenceValue) Kind() Kind {
	return ServiceReference
}

func (s ServiceReferenceValue) String() string {
	return ""
}

func (s ServiceReferenceValue) Boolean() bool {
	return false
}

func (s ServiceReferenceValue) Integer() int {
	return 0
}
