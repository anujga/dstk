package main

import (
	"errors"
	"fmt"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/bdb"
	"github.com/anujga/dstk/pkg/ss"
)

// 2. Define the state for a given partition and implement ss.Consumer
type partitionConsumer struct {
	p  *dstk.Partition
	pc *bdb.Wrapper
}

func (m *partitionConsumer) GetSnapshot() ss.AppState {
	return nil
}

func (m *partitionConsumer) ApplySnapshot(as ss.AppState) error {
	if as.State() == nil {
		return nil
	}
	return errors.New("unexpected state")
}

func (m *partitionConsumer) Meta() *dstk.Partition {
	return m.p
}

// thread safe
func (m *partitionConsumer) get(req *dstk.DcGetReq) (interface{}, error) {
	return m.pc.Get(req.GetKey())
}

func (m *partitionConsumer) put(req *dstk.DcPutReq) (interface{}, error) {
	return nil, m.pc.Put(req.GetKey(), req.GetValue(), req.GetTtlSeconds())
}

func (m *partitionConsumer) remove(req *dstk.DcRemoveReq) (interface{}, error) {
	return nil, m.pc.Remove(req.GetKey())
}

/// this method does not have to be thread safe
func (m *partitionConsumer) Process(msg0 ss.Msg) (interface{}, error) {
	msg := msg0.(*DcRequest)
	request := msg.grpcRequest
	switch request.(type) {
	case *dstk.DcGetReq:
		return m.get(request.(*dstk.DcGetReq))
	case *dstk.DcPutReq:
		return m.put(request.(*dstk.DcPutReq))
	case *dstk.DcRemoveReq:
		return m.remove(request.(*dstk.DcRemoveReq))
	default:
		return nil, errors.New(fmt.Sprintf("invalid message %v", msg))
	}
}
