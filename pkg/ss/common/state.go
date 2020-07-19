package common

type AppState interface {
	State() interface{}
}

type AppStateImpl struct {
	S interface{}
}

func (a *AppStateImpl) ResponseChannel() chan interface{} {
	return nil
}

func (a *AppStateImpl) State() interface{} {
	return a.S
}
