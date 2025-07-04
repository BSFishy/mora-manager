package state

type HasState interface {
	GetState() *State
}
