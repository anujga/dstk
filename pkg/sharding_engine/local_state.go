package se

import (
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/rangemap"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"sync/atomic"
)

type state struct {
	rangeMap     *rangemap.RangeMap
	pbs          []*pb.Partition
	lastModified int64
}

type stateHolder struct {
	r atomic.Value
}

func (s *stateHolder) Clear() {
	a := state{
		rangeMap: rangemap.New(2),
		pbs:      []*pb.Partition{},
	}
	s.r.Store(&a)
}

func (s *stateHolder) Parts() ([]*pb.Partition, error) {
	a := s.r.Load()
	if a == nil {
		return nil, core.ErrInfo(
			codes.Internal,
			"Reading partitions from uninitialized cache",
			"s", s).Err()
	}
	return a.(*state).pbs, nil
}

func (s *stateHolder) LastModified() int64 {
	a := s.r.Load()
	if a == nil {
		return 0
	}
	return a.(*state).lastModified
}

func (s *stateHolder) Get(key core.KeyT) (*pb.Partition, error) {
	state := s.r.Load().(*state)
	p, err := state.rangeMap.Get(key)
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
func (s *stateHolder) UpdateTree(parts *pb.Partitions) error {
	t := rangemap.New(16) // log(100K) expected count of partition

	for _, p := range parts.GetParts() {
		err := t.Put(&PartRange{p: p})
		if err != nil {
			return err
		}
	}

	zap.S().Infow("Partitions found", "count", len(parts.GetParts()))

	s.r.Store(&state{
		rangeMap:     t,
		pbs:          parts.GetParts(),
	})

	return nil
}
