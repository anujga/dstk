package main

import (
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
)

type DcRequest struct {
	grpcRequest interface{}
	C           chan interface{}
}

func (r *DcRequest) ResponseChannel() chan interface{} {
	return r.C
}

func (r *DcRequest) ReadOnly() bool {
	_, ok := r.grpcRequest.(*dstk.DcGetReq)
	return ok
}

func (r *DcRequest) Key() core.KeyT {
	switch r.grpcRequest.(type) {
	case *dstk.DcGetReq:
		return r.grpcRequest.(*dstk.DcGetReq).GetKey()
	case *dstk.DcPutReq:
		return r.grpcRequest.(*dstk.DcPutReq).GetKey()
	case *dstk.DcRemoveReq:
		return r.grpcRequest.(*dstk.DcRemoveReq).GetKey()
	}
	return nil
}
