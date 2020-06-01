package main

import (
	"fmt"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/ss"
)

// 2. Define the state for a given partition and implement ss.PartHandler
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
	c := msg.ResponseChannel()
	if err == nil {
		c <- "counter incremented"
	} else {
		c <- err
	}
	close(c)
	return err == nil
}

// 3. implement ss.ConsumerFactory

type partitionCounterMaker struct {
	dbPathPrefix   string
	maxOutstanding int
}

func (m *partitionCounterMaker) getDbPath(p *dstk.Partition) string {
	return fmt.Sprintf("%s/%d", m.dbPathPrefix, p.GetId())
}

func (m *partitionCounterMaker) Make(p *dstk.Partition) (ss.PartHandler, int, error) {
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
