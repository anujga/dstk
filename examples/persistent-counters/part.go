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
	//go func() {
		msg := msg0.(*Request)
		var err error
		c := msg.ResponseChannel()
		// TODO better way to model get/inc requests
		if msg.V == 0 {
			if val, err := m.pc.Get(msg.K); err == nil {
				c <- val
			} else {
				c <- err
			}
		} else {
			err = m.pc.Inc(msg.K, msg.V)
			if err == nil {
				c <- "counter incremented"
			} else {
				c <- err
			}
		}
		close(c)
	//}()
	return true
}

// 3. implement ss.ConsumerFactory

type partitionCounterMaker struct {
	dbPathPrefix   string
	maxOutstanding int
}

func (m *partitionCounterMaker) getDbPath(p *dstk.Partition) string {
	return fmt.Sprintf("%s/%d", m.dbPathPrefix, p.GetId())
}

func (m *partitionCounterMaker) Make(p *dstk.Partition) (ss.Consumer, int, error) {
	// TODO: gracefully stop the db too
	pc, err := NewCounter(m.getDbPath(p))
	if err != nil {
		return nil, 0, err
	}
	return &partitionCounter{
		p:  p,
		pc: pc,
	}, m.maxOutstanding, nil
}
