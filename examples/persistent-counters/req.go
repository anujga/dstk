package main

import "github.com/anujga/dstk/pkg/ss"

type Request struct {
	K string
	V int64
	C chan interface{}
}

func (r *Request) ResponseChannel() chan interface{} {
	return r.C
}

func (r *Request) ReadOnly() bool {
	return false
}

func (r *Request) Key() ss.KeyT {
	return []byte(r.K)
}
