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

func (m *partitionConsumer) Meta() *dstk.Partition {
	return m.p
}

// thread safe
func (m *partitionConsumer) get(req *dstk.DcGetReq, ch chan interface{}) bool {
	if val, err := m.pc.Get(req.GetKey()); err == nil {
		ch <- val
		return true
	} else {
		ch <- err
		return false
	}
}

func (m *partitionConsumer) put(req *dstk.DcPutReq, ch chan interface{}) bool {
	if err := m.pc.Put(req.GetKey(), req.GetValue(), req.GetTtlSeconds()); err == nil {
		return true
	} else {
		ch <- err
		return false
	}
}

func (m *partitionConsumer) remove(req *dstk.DcRemoveReq, ch chan interface{}) bool {
	if err := m.pc.Remove(req.GetKey()); err == nil {
		return true
	} else {
		ch <- err
		return false
	}
}

/// this method does not have to be thread safe
func (m *partitionConsumer) Process(msg0 ss.Msg) bool {
	msg := msg0.(*DcRequest)
	c := msg0.ResponseChannel()
	defer close(c)
	switch msg.grpcRequest.(type) {
	case *dstk.DcGetReq:
		return m.get(msg.grpcRequest.(*dstk.DcGetReq), msg.C)
	case *dstk.DcPutReq:
		return m.put(msg.grpcRequest.(*dstk.DcPutReq), msg.C)
	case *dstk.DcRemoveReq:
		return m.remove(msg.grpcRequest.(*dstk.DcRemoveReq), msg.C)
	default:
		c <- errors.New(fmt.Sprintf("invalid message %v", msg))
		return false
	}
}
