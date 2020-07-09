package ss

import (
	"context"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/rangemap"
	se "github.com/anujga/dstk/pkg/sharding_engine"
	"go.uber.org/zap"
	"sync/atomic"
	"time"
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

type state struct {
	m            *rangemap.RangeMap
	lastModified int64
}

type appState struct {
	s interface{}
}

func (a *appState) ResponseChannel() chan interface{} {
	return nil
}

func (a *appState) State() interface{} {
	return a.s
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

func (pm *PartitionMgr) ResetMap(s *state) {
	old := pm.State()
	pm.state.Store(s)
	//todo: we need to close old gracefully. correct algorithm
	//would be to ref count state and then close. here we just
	// sleep for 1 minute
	go func() {
		<-time.NewTimer(1 * time.Minute).C
		if old != nil {
			CloseConsumers(old.m)
		}
	}()

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
	if newTime <= s.lastModified {
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
	}

	for _, p := range rs.GetParts() {
		err := s.add(p, pm.slog, pm.consumer, pm.initStateMaker)
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

// path=control
func (s *state) add(p *pb.Partition, slog *zap.SugaredLogger, consumer ConsumerFactory, initStateMaker func() interface{}) error {
	var err error
	slog.Info("AddPartition Start", "part", p)
	defer slog.Info("AddPartition Status", "part", p, "err", err)
	c, maxOutstanding, err := consumer.Make(p)
	if err != nil {
		return err
	}
	part := NewPartRange(p, c, maxOutstanding)
	if err = s.m.Put(part); err != nil {
		return err
	}
	part.Mailbox() <- &appState{s: initStateMaker()}
	part.Run()
	return nil
}
