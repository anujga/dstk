package ss

import (
	"context"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/rangemap"
	se "github.com/anujga/dstk/pkg/sharding_engine"
	"go.uber.org/zap"
	"sync/atomic"
)

// todo: this is the 4th implementation of the range map.
// need to define a proper data structure that can be reused
type PartitionMgr struct {
	consumer       ConsumerFactory
	state          atomic.Value //[state]
	rpc            pb.SeWorkerApiClient
	id             se.WorkerId
	slog           *zap.SugaredLogger
	initStateMaker func() interface{}
}

func (pm *PartitionMgr) Map() *rangemap.RangeMap {
	return pm.State().m
}

func (pm *PartitionMgr) State() *state {
	s := pm.state.Load()
	if s == nil {
		return nil
	}
	return s.(*state)
}

func (pm *PartitionMgr) ResetMap(s *state) *state {
	old := pm.State()
	pm.state.Store(s)
	return old
}

func CloseConsumers(rs *rangemap.RangeMap) {
	for i := range rs.Iter(core.MinKey) {
		p := i.(*PartRange)
		//the consumer will continue to work till mailbox is empty
		p.Stop()
	}
}

// Path = data
func (pm *PartitionMgr) Find(key core.KeyT) (PartitionActor, error) {
	if rng, err := pm.Map().Get(key); err == nil {
		return rng.(PartitionActor), nil
	} else {
		return nil, err
	}
}

//todo: should indicate whether changes were applied or not
// poor algorithm that creates a new map.
//only apply delta changes. linear serializability is invalid here
//because there can be 2 mailbox for a given partition
func (pm *PartitionMgr) syncSe() error {
	rs, err := pm.rpc.MyParts(context.TODO(),
		&pb.MyPartsReq{WorkerId: int64(pm.id)})
	if err != nil {
		return err
	}

	newTime := rs.GetLastModified()
	s := pm.State()
	if s != nil && newTime <= s.lastModified {
		return nil
	}
	newState, err := newMap(pm, rs)
	if err != nil {
		return err
	}

	pm.ResetMap(newState)
	return nil
}

func newMap(pm *PartitionMgr, rs *pb.PartList) (*state, error) {
	s := &state{
		m:            rangemap.New(15),
		lastModified: rs.GetLastModified(),
		logger:       pm.slog,
	}

	for _, p := range rs.GetParts() {
		part, err := s.add(p, pm.consumer, nil)
		if err != nil {
			return nil, err
		}
		part.Mailbox() <- &appState{s: pm.initStateMaker()}
	}
	return s, nil
}
