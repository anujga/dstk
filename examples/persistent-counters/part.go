package main

import (
	"fmt"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/ss"
)

// 2. Define the state for a given partition and implement ss.Consumer
type partitionCounter struct {
	p  *dstk.Partition
	pc *PersistentCounter
}

func (m *partitionCounter) Meta() *dstk.Partition {
	return m.p
}

/// this method does not have to be thread safe
func (m *partitionCounter) Process(msg0 ss.Msg) bool {
	msg := msg0.(*Request)
	err := m.pc.Inc(msg.K, msg.V)
	//newVal, err := m.pc.Get(msg.K)
	//fmt.Printf("new value  %d\n", newVal)
	return err == nil
}

// 3. implement ss.ConsumerFactory

type partitionCounterMaker struct {
	maxOutstanding int
}

func (m *partitionCounterMaker) Make(p *dstk.Partition) (ss.Consumer, int, error) {
	// TODO: gracefully stop the db too
	pc, err := NewCounter(fmt.Sprintf("/var/tmp/counter-db/%d", p.GetId()))
	if err != nil {
		return nil, 0, err
	}
	return &partitionCounter{
		p:  p,
		pc: pc,
	}, m.maxOutstanding, nil
}
