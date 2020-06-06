package main

import (
	"github.com/anujga/dstk/pkg/core"
)

type RequestType byte

const (
	Get RequestType = iota
	Inc
	Remove
)

// TODO looks a bit odd to use same request for all
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

func (r *Request) Key() core.KeyT {
	return []byte(r.K)
}

func newIncRequest(key string, value int64, ttlSeconds float64, ch chan interface{}) *Request {
	return &Request{
		K:           key,
		V:           value,
		C:           ch,
		TtlSeconds:  ttlSeconds,
		RequestType: Inc,
	}
}

func newGetRequest(key string, ch chan interface{}) *Request {
	return &Request{
		K:           key,
		C:           ch,
		RequestType: Get,
	}
}

func newRemoveRequest(key string, ch chan interface{}) *Request {
	return &Request{
		K:           key,
		C:           ch,
		RequestType: Remove,
	}
}