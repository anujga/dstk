package se

import (
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/rangemap"
	"go.uber.org/zap"
	"sync/atomic"
)

type state struct {
	m            *rangemap.RangeMap
	pbs          []*pb.Partition
	lastModified int64
}

type stateHolder struct {
	r atomic.Value
}

func (s *stateHolder) Clear() {
	a := state{
		m:   rangemap.New(1),
		pbs: []*pb.Partition{},
	}
	s.r.Store(a)
}

func (s *stateHolder) Parts() ([]*pb.Partition, error) {
	return s.r.Load().(*state).pbs, nil
}

func (s *stateHolder) LastModified() int64 {
	return s.r.Load().(*state).lastModified
}

func (s *stateHolder) Get(key core.KeyT) (*pb.Partition, error) {
	m := s.r.Load().(*state)
	p, err := m.m.Get(key)
	if err != nil {
		return nil, err
	}

	part := p.(*PartRange)
	return part.p, nil
}

type PartRange struct {
	p *pb.Partition
}

func (x *PartRange) Start() core.KeyT {
	return x.p.GetStart()
}

func (x *PartRange) End() core.KeyT {
	return x.p.GetEnd()
}

//todo: use CAS on instead of blind replace to avoid lost update
// thread: unsafe. its actually safe but the above statement forces
// callers to call sequentially
func (s *stateHolder) UpdateTree(parts []*pb.Partition, lastModified int64) error {
	t := rangemap.New(16) // log(100K) expected count of partition

	for _, p := range parts {
		err := t.Put(&PartRange{p: p})
		if err != nil {
			return err
		}
	}

	zap.S().Infow("Partitions found", "count", len(parts))

	s.r.Store(&state{
		m:            t,
		pbs:          parts,
		lastModified: lastModified,
	})

	return nil
}
