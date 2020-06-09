package main

import (
	"github.com/anujga/dstk/pkg/core"
)

type RequestType byte

const (
	Get RequestType = iota
	Put
	Remove
)

type DcRequest struct {
	K core.KeyT
	V core.KeyT
	C chan interface{}
	TtlSeconds float64
	RequestType RequestType
}

func (r *DcRequest) ResponseChannel() chan interface{} {
	return r.C
}

func (r *DcRequest) ReadOnly() bool {
	return r.RequestType == Get
}

func (r *DcRequest) Key() core.KeyT {
	return r.K
}

func newPutRequest(key, value core.KeyT, ttlSeconds float64, ch chan interface{}) *DcRequest {
	return &DcRequest{
		K:           key,
		V:           value,
		C:           ch,
		TtlSeconds:  ttlSeconds,
		RequestType: Put,
	}
}

func newGetRequest(key core.KeyT, ch chan interface{}) *DcRequest {
	return &DcRequest{
		K:           key,
		C:           ch,
		RequestType: Get,
	}
}

func newRemoveRequest(key core.KeyT, ch chan interface{}) *DcRequest {
	return &DcRequest{
		K:           key,
		C:           ch,
		RequestType: Remove,
	}
}
