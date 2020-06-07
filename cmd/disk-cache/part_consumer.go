package main

import (
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

func (m *partitionConsumer) get(req *DcRequest) bool {
	if val, err := m.pc.Get(req.K); err == nil {
		req.C <- val
		return true
	} else {
		req.C <- err
		return false
	}
}

/// this method does not have to be thread safe
func (m *partitionConsumer) Process(msg0 ss.Msg) bool {
	msg := msg0.(*DcRequest)
	c := msg.ResponseChannel()
	defer close(c)
	switch msg.RequestType {
	case Get:
		return m.get(msg)
	}
	return true
}
