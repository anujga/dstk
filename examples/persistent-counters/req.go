package main

import "github.com/anujga/dstk/pkg/ss"

type RequestType byte

const (
	Get RequestType = iota
	Inc
	Remove
)

type Request struct {
	K string
	V int64
	C chan interface{}
	TtlSeconds float64
	RequestType RequestType
}

func (r *Request) ResponseChannel() chan interface{} {
	return r.C
}

func (r *Request) ReadOnly() bool {
	return r.RequestType == Get
}

func (r *Request) Key() ss.KeyT {
	return []byte(r.K)
}
