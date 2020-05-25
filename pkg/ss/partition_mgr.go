package ss

import (
	"bytes"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/google/btree"
	"go.uber.org/zap"
	"gopkg.in/errgo.v2/fmt/errors"
)

type PartItem struct {
	k        KeyT
	consumer Consumer
	mailBox  chan Msg
}

func (p *PartItem) Less(than btree.Item) bool {
	that := than.(*PartItem)
	e1, e2 := p.k, that.k
	return bytes.Compare(e1, e2) < 0
}

func (p *PartItem) Run() bool {
	for m := range p.mailBox {
		p.consumer.Process(m)
	}
	return true
}

//---

// todo: this is the 4th implementation of the range map.
// need to define a proper data structure that can be reused
type PartitionMgr struct {
	consumer ConsumerFactory
	partMap  *btree.BTree // PartItem
	lastPart *PartItem
	log      *zap.Logger
	slog     *zap.SugaredLogger
}

//todo: ensure there is at least 1 partition during construction
func NewPartitionMgr(consumer ConsumerFactory, log *zap.Logger) *PartitionMgr {
	return &PartitionMgr{
		consumer: consumer,
		partMap:  btree.New(32),
		lastPart: nil,
		log:      log,
		slog:     log.Sugar(),
	}
}

// Path = data
// Time = O(log N/ log 32), N ~ MAX_PARTITIONS ~ 1M
// Return value can only be nil if there are 0 partitions
func (m *PartitionMgr) Find(key KeyT) *PartItem {
	k := PartItem{k: key}

	var q = m.lastPart
	m.partMap.AscendGreaterOrEqual(&k, func(i btree.Item) bool {
		p := i.(*PartItem)
		q = p
		return false
	})

	return q
}

// path=control
func (m *PartitionMgr) Add(p *dstk.Partition) error {
	end := p.GetEnd()
	var err error

	m.slog.Info("AddPartition Start", "part", p)
	defer m.slog.Info("AddPartition Status", "part", p, "err", err)

	c, maxOutstanding := m.consumer.Make(p)
	part := PartItem{
		k:        end,
		consumer: c,
		mailBox:  make(chan Msg, maxOutstanding),
	}

	if len(end) == 0 {
		if m.lastPart != nil {
			err = errors.New("duplicate last partition")
			return err
		}
		m.lastPart = &part
	} else if nil != m.partMap.ReplaceOrInsert(&part) {
		err = errors.New("duplicate partition")
		return err
	}

	//todo: also check for valid start and other constraints

	go part.Run()

	return nil
}

// Single threaded router. 1 channel per partition
// path=data
func (m *PartitionMgr) OnMsg(msg Msg) error {
	p := m.Find(msg.Key())
	select {
	case p.mailBox <- msg:
		return nil
	default:
		return errors.Newf(
			"code=429. Partition Busy. Max outstanding allowed %d",
			cap(p.mailBox))
	}

}
