package ss

import (
	"bytes"
	"github.com/google/btree"
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
