package ss

import (
	"bytes"
	dstk "github.com/anujga/dstk/pkg/api/proto"
	"github.com/google/btree"
)

type ConsumerFactory interface {
	Make(p *dstk.Partition) Consumer
}

type PartItem struct {
	k        KeyT
	consumer Consumer
	q        chan Msg
}

func (p *PartItem) Less(than btree.Item) bool {
	that := than.(*PartItem)
	e1, e2 := p.consumer.Meta().GetEnd(), that.consumer.Meta().GetEnd()
	return bytes.Compare(e1, e2) < 0
}

type PartitionMgr struct {
}

// Single threaded router. 1 channel per partition
type STRouter struct {
	consumer ConsumerFactory
	partMap  *btree.BTree // PartItem
	lastPart *PartItem
}

func (r *STRouter) OnMsg(m Msg) {
	k := PartItem{k: m.Key()}

	var q = r.lastPart.q

	r.partMap.AscendGreaterOrEqual(&k, func(i btree.Item) bool {
		p := i.(*PartItem)
		q = p.q
		return false
	})

	q <- m
}
