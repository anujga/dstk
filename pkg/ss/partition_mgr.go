package ss

import (
	"context"
	pb "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/core"
	"github.com/anujga/dstk/pkg/rangemap"
	se "github.com/anujga/dstk/pkg/sharding_engine"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"sync/atomic"
	"time"
)

// todo: this is the 4th implementation of the range map.
// need to define a proper data structure that can be reused
type PartitionMgr struct {
	consumer ConsumerFactory
	state    atomic.Value //[state]
	rpc      pb.SeWorkerApiClient
	id       se.WorkerId
	slog     *zap.SugaredLogger
}

type state struct {
	m            *rangemap.RangeMap
	lastModified int64
}

func (pm *PartitionMgr) Map() *rangemap.RangeMap {
	return pm.State().m
}

func (pm *PartitionMgr) State() *state {
	return pm.state.Load().(*state)
}

func (pm *PartitionMgr) ResetMap(s *state) {
	old := pm.State()
	pm.state.Store(s)
	//todo: we need to close old gracefully. correct algorithm
	//would be to ref count state and then close. here we just
	// sleep for 1 minute
	go func() {
		<-time.NewTimer(1 * time.Minute).C
		old.m.Close()
	}()

}

// Path = data
func (pm *PartitionMgr) Find(key core.KeyT) (*PartRange, error) {
	if rng, err := pm.Map().Get(key); err == nil {
		return rng.(*PartRange), nil
	} else {
		return nil, err
	}
}

//todo: ensure there is at least 1 partition during construction
func NewPartitionMgr2(workerId se.WorkerId, consumer ConsumerFactory, rpc pb.SeWorkerApiClient) *PartitionMgr {
	pm := &PartitionMgr{
		consumer: consumer,
		rpc:      rpc,
		id:       workerId,
		slog:     zap.S().With("workerId", workerId),
	}

	core.Repeat(30*time.Second, func(timestamp time.Time) bool {
		err := pm.syncSe()
		if err != nil {
			pm.slog.Errorw("fetch updates from SE",
				"err", err)
		} else {
			delay := timestamp.UnixNano() - pm.State().lastModified
			pm.slog.Infow("fetch updates from SE",
				"time", timestamp,
				"delay", delay)
		}
		return true
	})
	return pm
}

//todo: should indicate whether changes were applied or not
func (pm *PartitionMgr) syncSe() error {
	rs, err := pm.rpc.MyParts(context.TODO(),
		&pb.MyPartsReq{WorkerId: int64(pm.id)})
	if err != nil {
		return err
	}

	newTime := rs.GetLastModified()
	if newTime <= pm.State().lastModified {
		return nil
	}
	s, err := newMap(pm, rs)
	if err != nil {
		return err
	}

	pm.ResetMap(s)
	return nil
}

func newMap(pm *PartitionMgr, rs *pb.PartList) (*state, error) {
	s := &state{
		m:            rangemap.New(15),
		lastModified: rs.GetLastModified(),
	}

	for _, p := range rs.GetParts() {
		err := s.add(p, pm.slog, pm.consumer)
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

// path=control
func (s *state) add(p *pb.Partition, slog *zap.SugaredLogger, consumer ConsumerFactory) error {
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

	part.Run()
	return nil
}

// Single threaded router. 1 channel per partition
// path=data
func (pm *PartitionMgr) OnMsg(msg Msg) error {
	p, err := pm.Find(msg.Key())
	if err != nil {
		return err
	}
	select {
	case p.mailBox <- msg:
		return nil
	default:
		return core.ErrInfo(codes.ResourceExhausted, "Partition Busy",
			"capacity", cap(p.mailBox),
			"partition", p.Id()).Err()

	}
}
