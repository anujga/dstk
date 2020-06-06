package ss

import (
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/anujga/dstk/pkg/rangemap"
	"go.uber.org/zap"
	"gopkg.in/errgo.v2/fmt/errors"
)

// todo: this is the 4th implementation of the range map.
// need to define a proper data structure that can be reused
type PartitionMgr struct {
	consumer ConsumerFactory
	rangeMap *rangemap.RangeMap
	log      *zap.Logger
	slog     *zap.SugaredLogger
}

//todo: ensure there is at least 1 partition during construction
func NewPartitionMgr(consumer ConsumerFactory, log *zap.Logger) *PartitionMgr {
	return &PartitionMgr{
		consumer: consumer,
		rangeMap: rangemap.New(32),
		log:      log,
		slog:     log.Sugar(),
	}
}

// Path = data
func (pm *PartitionMgr) Find(key KeyT) (*PartRange, error) {
	if rng, err := pm.rangeMap.Get(key); err == nil {
		return rng.(*PartRange), nil
	} else {
		return nil, err
	}
}

// path=control
func (pm *PartitionMgr) Add(p *dstk.Partition) error {
	var err error
	pm.slog.Info("AddPartition Start", "part", p)
	defer pm.slog.Info("AddPartition Status", "part", p, "err", err)
	c, maxOutstanding, err := pm.consumer.Make(p)
	if err != nil {
		return err
	}
	part := &PartRange{
		partition: p,
		consumer:  c,
		mailBox:   make(chan Msg, maxOutstanding),
	}
	if err := pm.rangeMap.Put(part); err == nil {
		//todo: also check for valid start and other constraints
		go part.Run()
		return nil
	} else {
		return err
	}
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
		return errors.Newf(
			"code=429. Partition Busy. Max outstanding allowed %d",
			cap(p.mailBox))
	}
}
