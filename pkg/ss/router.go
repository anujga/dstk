package ss

import (
	"bytes"
	"errors"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/google/btree"
	"go.uber.org/zap"
)

type ConsumerFactory interface {
	Make(p *dstk.Partition) Consumer
}

type PartItem struct {
	k        KeyT
	consumer Consumer
	mailBox  chan Msg
}

func (p *PartItem) Less(than btree.Item) bool {
	that := than.(*PartItem)
	e1, e2 := p.consumer.Meta().GetEnd(), that.consumer.Meta().GetEnd()
	return bytes.Compare(e1, e2) < 0
}

// this is the 4rth implementation of the range map.
// need to define a proper data structure that can be reused
type PartitionMgr struct {
	consumer ConsumerFactory
	partMap  *btree.BTree // PartItem
	lastPart *PartItem
	log      *zap.Logger
	slog     *zap.SugaredLogger
}

// path=data
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

// path=control. performance not so critical
func (m *PartitionMgr) Add(p *dstk.Partition) error {
	end := p.GetEnd()
	var err error

	m.slog.Info("AddPartition Start", "end", end)
	defer m.slog.Info("AddPartition Status", "end", end, "err", err)

	if end == nil {
		if m.lastPart != nil {
			err = errors.New("duplicate last partition")
			return err
		}

	}
	c := m.consumer.Make(p)
	i := PartItem{
		k:        end,
		consumer: c,
		mailBox:  make(chan Msg, c.MaxOutstanding()),
	}
	if nil != m.partMap.ReplaceOrInsert(&i) {
		err = errors.New("duplicate partition")
		return err
	}

	//todo: also check for valid start and other constraints
	return nil
}

// Single threaded router. 1 channel per partition
type STRouter struct {
	parts *PartitionMgr
}

func (s *STRouter) OnMsg(m Msg) {
	p := s.parts.Find(m.Key())
	p.mailBox <- m
}
