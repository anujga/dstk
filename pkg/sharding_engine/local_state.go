package se

import (
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/rangemap"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync/atomic"
)

type state struct {
	rangeMap     rangemap.RangeMap
	pbs          []*pb.Partition
	lastModified int64
}

type partitionMgr struct {
	r atomic.Value
}

func (s *partitionMgr) Clear() {
	a := state{
		rangeMap: rangemap.NewBtreeRange(2),
		pbs:      []*pb.Partition{},
	}
	s.r.Store(&a)
}

func (s *partitionMgr) Parts() ([]*pb.Partition, error) {
	a := s.r.Load()
	if a == nil {
		return nil, core.ErrInfo(
			codes.Internal,
			"Reading partitions from uninitialized partitionCache",
			"s", s).Err()
	}
	return a.(*state).pbs, nil
}

func (s *partitionMgr) LastModified() int64 {
	a := s.r.Load()
	if a == nil {
		return 0
	}
	return a.(*state).lastModified
}

func (s *partitionMgr) Get(key core.KeyT) (*pb.Partition, *status.Status) {
	state := s.r.Load().(*state)
	p, found, err := state.rangeMap.Get(key)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, core.ErrInfo(codes.InvalidArgument,
			"key does not belong to this paritionMap",
			"key", key,
		)
	}

	part := p.(*PartRange)
	return part.p, nil
}

//todo: use CAS on instead of blind replace to avoid lost update
// thread: unsafe. its actually safe but the above statement forces
// callers to call sequentially
func (s *partitionMgr) UpdateTree(parts *pb.Partitions, lastModified int64) error {
	t := rangemap.NewBtreeRange(16) // log(100K) expected count of partition

	for _, p := range parts.GetParts() {
		err := t.Put(&PartRange{p: p})
		if err != nil {
			return err.Err()
		}
	}

	zap.S().Infow("Partitions found", "count", len(parts.GetParts()))

	s.r.Store(&state{
		rangeMap:     t,
		pbs:          parts.GetParts(),
		lastModified: lastModified,
	})

	return nil
}
