package core

import "io"

type FutureErr struct {
	done   chan error
	closer io.Closer

	io.Closer
}

func (f *FutureErr) Close() error {
	return f.closer.Close()
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

func RunAsync2(fn func() error, close io.Closer) *FutureErr {
	return &FutureErr{
		done:   make(chan error, 1),
		closer: close,
	}
}
