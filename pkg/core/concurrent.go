package core

type FutureErr struct {
	done chan error
}

func (f *FutureErr) Done() <-chan error {
	return f.done
}

func (f *FutureErr) Wait() error {
	return <-f.done
}

func (f *FutureErr) Cancel() <-chan error {
	panic("not implemented")
}

func (f *FutureErr) Complete(fn func() error) *FutureErr {
	go func() {
		f.done <- fn()
	}()
	return f
}

func NewPromise() *FutureErr {
	return &FutureErr{
		done: make(chan error, 1),
	}
}

func RunAsync(fn func() error) *FutureErr {
	return NewPromise().Complete(fn)
}
