package core

type HasUser interface {
	GetUser() string
}

type HasEnvironment interface {
	GetEnvironment() string
}
